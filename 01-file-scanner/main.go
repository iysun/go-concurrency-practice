package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
)

// Result holds stats for a single file
type Result struct {
	Path string
	Size int64
}

// Scanner walks a directory tree concurrently.
// TODO: implement a worker pool to limit goroutine count
type Scanner struct {
	root        string
	workerCount int
	jobs        chan string  // directories to scan
	results     chan Result  // scanned file results
	wg          sync.WaitGroup
}

func NewScanner(root string, workers int) *Scanner {
	return &Scanner{
		root:        root,
		workerCount: workers,
		jobs:        make(chan string, 100),
		results:     make(chan Result, 100),
	}
}

// worker reads directories from s.jobs, lists entries, and either
// recurses into subdirs or sends file results.
// TODO: implement this
func (s *Scanner) worker() {
	defer s.wg.Done()
	for dir := range s.jobs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			full := filepath.Join(dir, e.Name())
			if e.IsDir() {
				// TODO: send full into s.jobs (be careful: this can block if channel is full)
				_ = full
			} else {
				// TODO: get file info and send Result into s.results
			}
		}
	}
}

// Scan starts workers, seeds the root directory, and collects results.
// TODO: implement graceful shutdown via context
func (s *Scanner) Scan() (totalFiles int64, totalSize int64) {
	// Start workers
	for i := 0; i < s.workerCount; i++ {
		s.wg.Add(1)
		go s.worker()
	}

	// Seed the root
	s.jobs <- s.root

	// Close jobs when all workers finish
	go func() {
		s.wg.Wait()
		close(s.results)
	}()

	// Collect results
	for r := range s.results {
		atomic.AddInt64(&totalFiles, 1)
		atomic.AddInt64(&totalSize, r.Size)
		_ = r
	}
	return
}

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}

	scanner := NewScanner(root, 8)
	files, size := scanner.Scan()
	fmt.Printf("Files: %d  Total size: %d bytes\n", files, size)
}
