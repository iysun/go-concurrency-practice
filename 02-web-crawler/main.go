package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Crawler performs concurrent BFS link crawling within a single domain.
type Crawler struct {
	baseURL     string
	maxDepth    int
	concurrency int
	visited     sync.Map     // map[string]bool — concurrent-safe visited set
	jobs        chan job
	results     chan string   // collected URLs
	client      *http.Client
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
// TODO: implement using net/http and golang.org/x/net/html (or regexp as fallback)
func (c *Crawler) fetch(pageURL string) ([]string, error) {
	// hint: resp, err := c.client.Get(pageURL)
	// hint: parse <a href="..."> tags from resp.Body
	return nil, fmt.Errorf("not implemented")
}

// sameDomain returns true if link belongs to the same host as c.baseURL.
func (c *Crawler) sameDomain(link string) bool {
	base, err := url.Parse(c.baseURL)
	if err != nil {
		return false
	}
	u, err := url.Parse(link)
	if err != nil {
		return false
	}
	return u.Host == base.Host
}

// worker processes jobs from c.jobs, fetches the page, and enqueues new links.
// TODO: implement
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
			if j.depth > c.maxDepth {
				continue
			}
			// TODO: skip already-visited URLs (sync.Map)
			// TODO: fetch links and enqueue unseen same-domain ones
			_ = j
		}
	}
}

// Crawl seeds the root URL and fans out workers.
// TODO: implement graceful shutdown when jobs channel drains
func (c *Crawler) Crawl(ctx context.Context) []string {
	var wg sync.WaitGroup
	for i := 0; i < c.concurrency; i++ {
		wg.Add(1)
		go c.worker(ctx, &wg)
	}

	c.jobs <- job{url: c.baseURL, depth: 0}

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

	crawler := NewCrawler("https://example.com", 3, 8)
	links := crawler.Crawl(ctx)
	fmt.Printf("Found %d links\n", len(links))
	for _, l := range links {
		fmt.Println(" ", l)
	}
}
