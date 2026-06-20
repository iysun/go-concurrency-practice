package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// Crawler performs concurrent BFS link crawling within a single domain.
type Crawler struct {
	baseURL     string
	maxDepth    int
	concurrency int
	visited     sync.Map // map[string]bool — concurrent-safe visited set
	jobs        chan job
	results     chan string // collected URLs
	client      *http.Client
	pending     sync.WaitGroup
}

type job struct {
	url   string
	depth int
}

func NewCrawler(baseURL string, maxDepth, concurrency int) *Crawler {
	return &Crawler{
		baseURL:     baseURL,
		maxDepth:    maxDepth,
		concurrency: concurrency,
		jobs:        make(chan job, 256),
		results:     make(chan string, 256),
		client:      &http.Client{Timeout: 10 * time.Second},
	}
}

// fetch downloads a page and returns all href links found on it.
func (c *Crawler) fetch(ctx context.Context, pageURL string) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch err: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status code %d for %s", resp.StatusCode, pageURL)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}
	links := extractLinks(string(body), pageURL)
	return links, nil
}

// worker processes jobs from c.jobs, fetches the page, and enqueues new links.
func (c *Crawler) worker(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case j, ok := <-c.jobs:
			if !ok {
				return
			}
			c.process(ctx, j)
		}
	}
}

// process handles a single job. It calls pending.Done exactly once on every
// exit path so the pending counter reaches zero when the frontier drains.
func (c *Crawler) process(ctx context.Context, j job) {
	defer c.pending.Done()

	if j.depth > c.maxDepth {
		return
	}
	if _, loaded := c.visited.LoadOrStore(j.url, 1); loaded {
		return
	}
	c.results <- j.url

	links, err := c.fetch(ctx, j.url)
	if err != nil {
		log.Printf("fetch %s: %v", j.url, err)
		return
	}
	for _, l := range links {
		c.enqueue(ctx, job{url: l, depth: j.depth + 1})
	}
}

func (c *Crawler) enqueue(ctx context.Context, j job) {
	c.pending.Add(1)
	select {
	case c.jobs <- j:
	case <-ctx.Done():
		// ctx 取消时发送可能永久阻塞（buffer 满 + worker 已退出），
		// 这里走 ctx 分支并撤销刚才的 Add，避免 pending 计数泄漏。
		c.pending.Done()
	}
}

// Crawl seeds the root URL and fans out workers.
func (c *Crawler) Crawl(ctx context.Context) []string {
	var wg sync.WaitGroup
	for i := 0; i < c.concurrency; i++ {
		wg.Add(1)
		go c.worker(ctx, &wg)
	}

	c.enqueue(ctx, job{url: c.baseURL, depth: 0})

	go func() {
		c.pending.Wait()
		close(c.jobs)
	}()

	go func() {
		wg.Wait()
		close(c.results)
	}()

	var found []string
	for u := range c.results {
		found = append(found, u)
	}
	return found
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	crawler := NewCrawler("https://www.baidu.com", 30, 8)
	links := crawler.Crawl(ctx)
	fmt.Printf("Found %d links\n", len(links))
	for _, l := range links {
		fmt.Println(" ", l)
	}
}
