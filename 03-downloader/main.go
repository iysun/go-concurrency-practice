package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// var defaultChunks = runtime.NumCPU()
var defaultChunks = 32

// Chunk represents one slice of the file to download.
type Chunk struct {
	index int
	start int64
	end   int64
}

// Progress tracks download progress across chunks.
// downloaded is updated on the hot path via atomics (no lock, no I/O);
// a separate ticker goroutine reads it to print at a fixed interval.
type Progress struct {
	downloaded atomic.Int64
	totalSize  int64
}

func (p *Progress) add(n int64) {
	p.downloaded.Add(n)
}

func (p *Progress) print() {
	pct := float64(p.downloaded.Load()) / float64(p.totalSize) * 100
	fmt.Printf("\rProgress: %.1f%%", pct)
}

// Downloader fetches a file in parallel chunks using HTTP Range requests.
type Downloader struct {
	url       string
	dest      string
	numChunks int
	client    *http.Client
}

func NewDownloader(url, dest string, chunks int) *Downloader {
	// Let the transport keep one idle connection per concurrent chunk so
	// connections are reused instead of being closed and re-dialed.
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.MaxIdleConns = chunks
	tr.MaxIdleConnsPerHost = chunks
	return &Downloader{
		url:       url,
		dest:      dest,
		numChunks: chunks,
		client:    &http.Client{Transport: tr},
	}
}

// getContentLength sends a HEAD request to determine file size.
// Returns 0 if server doesn't support Range requests.
func (d *Downloader) getContentLength() (int64, error) {
	// hint: use http.MethodHead and check Accept-Ranges header
	req, err := http.NewRequest(http.MethodHead, d.url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	acceptRanges := resp.Header.Get("Accept-Ranges")
	supportsRange := acceptRanges == "bytes"

	if !supportsRange {
		return 0, fmt.Errorf("服务器不支持 Range 请求，无法进行分片下载")
	}
	contentLength := resp.Header.Get("Content-Length")

	if contentLength == "" {
		return 0, fmt.Errorf("服务器未返回 Content-Length")
	}

	fileSize, err := strconv.ParseInt(contentLength, 10, 64)
	if err != nil {
		return 0, err
	}

	fmt.Printf("文件大小: %d bytes (%.2f MB), 支持分片下载\n",
		fileSize, float64(fileSize)/(1024*1024))

	return fileSize, nil
}

// downloadChunk fetches bytes [c.start, c.end] and writes them
// into the correct offset of the output file.
func (d *Downloader) downloadChunk(ctx context.Context, c Chunk, f *os.File, p *Progress) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, d.url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", c.start, c.end))

	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("分片 %d 相应异常: %s", c.index, resp.Status)
	}

	// TeeReader duplicates data to progressWriter for tracking,
	// while io.Copy streams directly to the file.
	written, err := io.Copy(
		NewSectionWriter(f, c.start),
		io.TeeReader(resp.Body, &progressWriter{p: p}),
	)
	if err != nil {
		return err
	}

	expectedLen := c.end - c.start + 1
	if written != expectedLen {
		return fmt.Errorf("分片 %d 数据长度不匹配: 期望 %d, 实际 %d", c.index, expectedLen, written)
	}

	return nil
}

// Download splits the file into chunks and downloads them concurrently.
func (d *Downloader) Download(ctx context.Context) error {
	size, err := d.getContentLength()
	if err != nil {
		return err
	}

	f, err := os.Create(d.dest)
	if err != nil {
		return err
	}
	defer f.Close()

	// Pre-allocate file
	if err := f.Truncate(size); err != nil {
		return err
	}

	progress := &Progress{totalSize: size}
	chunkSize := size / int64(d.numChunks)

	// Print progress at a fixed interval instead of on every block,
	// keeping terminal I/O off the download hot path.
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				progress.print()
			case <-done:
				progress.print() // final 100%
				return
			}
		}
	}()

	var wg sync.WaitGroup
	errs := make(chan error, d.numChunks)

	for i := 0; i < d.numChunks; i++ {
		start := int64(i) * chunkSize
		end := start + chunkSize - 1
		if i == d.numChunks-1 {
			end = size - 1
		}
		wg.Add(1)
		go func(c Chunk) {
			defer wg.Done()
			if err := d.downloadChunk(ctx, c, f, progress); err != nil {
				errs <- err
			}
		}(Chunk{index: i, start: start, end: end})
	}

	wg.Wait()
	close(done)
	close(errs)
	fmt.Println()

	// Return first error if any
	for e := range errs {
		return e
	}
	return nil
}

// progressWriter only accumulates the byte count on the hot path.
// It performs no terminal I/O — printing is done by a ticker goroutine.
type progressWriter struct {
	p *Progress
}

func (pw *progressWriter) Write(b []byte) (int, error) {
	pw.p.add(int64(len(b)))
	return len(b), nil
}

// sectionWriter adapts os.File.WriteAt to io.Writer at a fixed offset.
type sectionWriter struct {
	f      *os.File
	offset int64
}

func (sw *sectionWriter) Write(b []byte) (int, error) {
	n, err := sw.f.WriteAt(b, sw.offset)
	sw.offset += int64(n)
	return n, err
}

// 这个其实可以使用 go 1.20+ 的 io.offsetWriter 替换
func NewSectionWriter(f *os.File, off int64) *sectionWriter {
	return &sectionWriter{f: f, offset: off}
}

// https://dl.testfile.cc/100mb.dat
// go run .\03-downloader\ https://dl.testfile.cc/100mb.dat 100mb.dat
func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: downloader <url> <output-file>")
		os.Exit(1)
	}
	ctx := context.Background()
	d := NewDownloader(os.Args[1], os.Args[2], defaultChunks)
	if err := d.Download(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	fmt.Println("Done.")
}
