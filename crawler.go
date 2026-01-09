package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type crawlResult struct {
	url    string
	status string
	kind   string
	size   int64
	links  []string
}

type crawler struct {
	baseUrl     *url.URL
	visited     sync.Map
	results     chan crawlResult
	jobs        chan string
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	active      sync.WaitGroup
	concurrency int
}

func newCrawler(baseUrl *url.URL, results chan crawlResult, concurrency int) *crawler {
	ctx, cancel := context.WithCancel(context.Background())
	return &crawler{
		baseUrl:     baseUrl,
		results:     results,
		jobs:        make(chan string, 10000),
		ctx:         ctx,
		cancel:      cancel,
		concurrency: concurrency,
	}
}

func (c *crawler) start() {
	for i := 0; i < c.concurrency; i++ {
		c.wg.Add(1)
		go c.worker()
	}

	// Monitor for completion
	go func() {
		c.active.Wait()
		// No more active jobs, close jobs to signal workers to stop if they are waiting for jobs
		c.results <- crawlResult{url: "__FINISHED__"}
	}()
}

func (c *crawler) stop() {
	c.cancel()
	c.wg.Wait()
}

func (c *crawler) worker() {
	defer c.wg.Done()
	for {
		select {
		case <-c.ctx.Done():
			return
		case target, ok := <-c.jobs:
			if !ok {
				return
			}
			res := c.doCrawl(target)

			select {
			case <-c.ctx.Done():
				c.active.Done()
				return
			case c.results <- res:
			}

			for _, link := range res.links {
				if _, loaded := c.visited.LoadOrStore(link, true); !loaded {
					c.active.Add(1)
					select {
					case <-c.ctx.Done():
						c.active.Done()
						return
					case c.jobs <- link:
					default:
						c.active.Done()
					}
				}
			}
			c.active.Done()
		}
	}
}

func (c *crawler) doCrawl(targetUrl string) crawlResult {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(targetUrl)
	if err != nil {
		return crawlResult{url: targetUrl, status: "Error", kind: "N/A", size: 0}
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return crawlResult{url: targetUrl, status: "Read Err", kind: "N/A", size: 0}
	}

	contentType := resp.Header.Get("Content-Type")
	kind := "Other"
	if strings.Contains(contentType, "text/html") {
		kind = "document"
	} else if strings.Contains(contentType, "text/css") {
		kind = "stylesheet"
	} else if strings.Contains(contentType, "javascript") {
		kind = "script"
	} else if strings.Contains(contentType, "font") {
		kind = "font"
	} else if strings.Contains(contentType, "image/png") {
		kind = "png"
	} else if strings.Contains(contentType, "image/gif") {
		kind = "gif"
	} else if strings.Contains(contentType, "image/jpeg") {
		kind = "jpeg"
	} else if strings.Contains(contentType, "image/svg+xml") {
		kind = "svg+xml"
	} else if strings.Contains(contentType, "x-icon") || strings.Contains(contentType, "vnd.microsoft.icon") {
		kind = "x-icon"
	} else if strings.Contains(contentType, "manifest+json") {
		kind = "manifest"
	}

	var links []string
	if kind == "document" {
		links = extractLinks(strings.NewReader(string(bodyBytes)), targetUrl, c.baseUrl)
	}

	return crawlResult{
		url:    targetUrl,
		status: fmt.Sprintf("%d", resp.StatusCode),
		kind:   kind,
		size:   int64(len(bodyBytes)),
		links:  links,
	}
}
