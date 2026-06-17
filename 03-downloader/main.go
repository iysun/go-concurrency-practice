package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
)

const defaultChunks = 4

// Chunk represents one slice of the file to download.
type Chunk struct {
	index int
	start int64
	end   int64
}

// Progress tracks download progress across chunks.
type Progress struct {
	mu          sync.Mutex
	downloaded  int64
	totalSize   int64
}

func (p *Progress) add(n int64) {
	p.mu.Lock()
	p.downloaded += n
	p.mu.Unlock()
}

func (p *Progress) print() {
	p.mu.Lock()
	pct := float64(p.downloaded) / float64(p.totalSize) * 100
	p.mu.Unlock()
	fmt.Printf("\rProgress: %.1f%%", pct)
}

// Downloader fetches a file in parallel chunks using HTTP Range requests.
type Downloader struct {
	url        string
	dest       string
	numChunks  int
	client     *http.Client
}

func NewDownloader(url, dest string, chunks int) *Downloader {
	return &Downloader{url: url, dest: dest, numChunks: chunks, client: &http.Client{}}
}

// getContentLength sends a HEAD request to determine file size.
// Returns 0 if server doesn't support Range requests.
// TODO: implement
func (d *Downloader) getContentLength() (int64, error) {
	// hint: use http.MethodHead and check Accept-Ranges header
	return 0, fmt.Errorf("not implemented")
}

// downloadChunk fetches bytes [c.start, c.end] and writes them
// into the correct offset of the output file.
// TODO: implement
func (d *Downloader) downloadChunk(ctx context.Context, c Chunk, f *os.File, p *Progress) error {
	// hint: set "Range: bytes=start-end" header
	// hint: use io.Copy with a progress-tracking writer
	_ = c
	_ = f
	_ = p
	return fmt.Errorf("not implemented")
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
	close(errs)
	fmt.Println()

	// Return first error if any
	for e := range errs {
		return e
	}
	return nil
}

// progressWriter wraps an io.Writer and reports bytes written to Progress.
type progressWriter struct {
	w io.Writer
	p *Progress
}

func (pw *progressWriter) Write(b []byte) (int, error) {
	n, err := pw.w.Write(b)
	pw.p.add(int64(n))
	pw.p.print()
	return n, err
}

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
