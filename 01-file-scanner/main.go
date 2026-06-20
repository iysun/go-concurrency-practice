package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// Result holds stats for a single file
type Result struct {
	Path string
	Size int64
}

// Scanner walks a directory tree concurrently.
type Scanner struct {
	root        string
	workerCount int
	jobs        chan string // directories to scan
	results     chan Result // scanned file results
	wg          sync.WaitGroup
	pending     sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewScanner(root string, workers int) *Scanner {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	return &Scanner{
		root:        root,
		workerCount: workers,
		jobs:        make(chan string, 100),
		results:     make(chan Result, 100),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// worker reads directories from s.jobs, lists entries, and either
// recurses into subdirs or sends file results.
func (s *Scanner) worker() {
	for dir := range s.jobs {
		s.processDir(dir)
		s.pending.Done()
	}
	s.wg.Done()
}

func (s *Scanner) enqueue(full string) {
	s.pending.Add(1)
	go func() {
		select {
		case s.jobs <- full:
		case <-s.ctx.Done():
			s.pending.Done()
		}
	}()
}

func (s *Scanner) processDir(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		full := filepath.Join(dir, e.Name())
		if e.IsDir() {
			s.enqueue(full)
		} else {
			info, err := e.Info()
			if err != nil {
				continue
			}
			select {
			case s.results <- Result{Path: full, Size: info.Size()}:
			case <-s.ctx.Done():
				return
			}
		}
	}

}

// Scan starts workers, seeds the root directory, and collects results.
func (s *Scanner) Scan() (totalFiles int64, totalSize int64) {
	// Start workers
	for i := 0; i < s.workerCount; i++ {
		s.wg.Add(1)
		go s.worker()
	}

	// Seed the root
	s.enqueue(s.root)

	// Close jobs when all workers finish
	go func() {
		s.pending.Wait()
		close(s.jobs)
	}()

	go func() {
		s.wg.Wait()
		close(s.results)
	}()

	// Collect results
	for r := range s.results {
		totalFiles++
		totalSize += r.Size
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
	fmt.Printf("Goroutines after scan: %d\n", runtime.NumGoroutine())
}
