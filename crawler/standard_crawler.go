package crawler

import (
	"context"
	"net/url"
	"sync"
)

// StandardCrawler is the default implementation of the Crawler interface
type StandardCrawler struct {
	collector   Collector
	registry    Registry
	concurrency int
	results     chan Resource
	jobs        chan string
	active      sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	baseURL     *url.URL
}

// NewStandardCrawler creates a new crawler instance
func NewStandardCrawler(collector Collector, registry Registry, concurrency int) *StandardCrawler {
	// Initialize with background context, will be replaced in Start
	ctx, cancel := context.WithCancel(context.Background())
	return &StandardCrawler{
		collector:   collector,
		registry:    registry,
		concurrency: concurrency,
		results:     make(chan Resource, 100),
		jobs:        make(chan string, 10000),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start begins the crawling process
func (c *StandardCrawler) Start(ctx context.Context, startURL string) error {
	u, err := url.Parse(startURL)
	if err != nil {
		return err
	}
	c.baseURL = u

	// Use the provided context for cancellation
	c.ctx, c.cancel = context.WithCancel(ctx)

	// Add start URL to jobs
	if c.registry.Visit(startURL) {
		c.active.Add(1)
		c.jobs <- startURL
	} else {
		// Should not happen if registry is fresh, but if reused...
		// If start URL is already visited, we might want to process it anyway?
		// Or just return?
		// For now, assume fresh registry.
		c.active.Add(1)
		c.jobs <- startURL
	}

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < c.concurrency; i++ {
		wg.Add(1)
		go c.worker(&wg)
	}

	// Wait for completion
	done := make(chan struct{})
	go func() {
		c.active.Wait()
		close(done)
	}()

	select {
	case <-c.ctx.Done():
		// Context cancelled
	case <-done:
		// All jobs finished
	}

	c.cancel() // Stop workers
	wg.Wait()
	close(c.results)

	return nil
}

// Stop halts the crawling process
func (c *StandardCrawler) Stop() {
	c.cancel()
}

// Results returns the channel of discovered resources
func (c *StandardCrawler) Results() <-chan Resource {
	return c.results
}

func (c *StandardCrawler) worker(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-c.ctx.Done():
			return
		case u, ok := <-c.jobs:
			if !ok {
				return
			}

			// Process the URL
			res, err := c.collector.Collect(c.ctx, u)
			if err != nil {
				// If resource is partial (e.g. error status), send it
				if res != nil {
					c.sendResult(*res)
				}
				c.active.Done()
				continue
			}

			// Send successful result
			c.sendResult(*res)

			// Process links
			for _, link := range res.Links {
				// Enforce same-domain policy
				parsedLink, err := url.Parse(link)
				if err != nil {
					continue
				}

				if parsedLink.Host == c.baseURL.Host {
					if c.registry.Visit(link) {
						c.active.Add(1)
						select {
						case c.jobs <- link:
						case <-c.ctx.Done():
							c.active.Done()
							return
						}
					}
				}
			}
			c.active.Done()
		}
	}
}

func (c *StandardCrawler) sendResult(res Resource) {
	select {
	case c.results <- res:
	case <-c.ctx.Done():
	}
}
