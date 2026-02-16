package crawler_test

import (
	"context"
	"testing"
	"time"

	"github.com/jturmel/huntsman/crawler"
)

func TestStandardCrawler_Start(t *testing.T) {
	// Setup mock collector
	collector := &MockCollector{}

	// Setup mock registry
	registry := &MockRegistry{}

	// Create crawler
	c := crawler.NewStandardCrawler(collector, registry, 1)

	// Start crawling
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() {
		if err := c.Start(ctx, "http://example.com"); err != nil {
			t.Errorf("Start failed: %v", err)
		}
	}()

	// Consume results
	select {
	case res := <-c.Results():
		if res.URL != "http://example.com" {
			t.Errorf("Expected URL http://example.com, got %s", res.URL)
		}
	case <-ctx.Done():
		t.Fatal("Timeout waiting for results")
	}

	c.Stop()
}

type MockCollectorWithLinks struct {
	Links map[string][]string
}

func (m *MockCollectorWithLinks) Collect(ctx context.Context, targetURL string) (*crawler.Resource, error) {
	links := m.Links[targetURL]
	return &crawler.Resource{
		URL:    targetURL,
		Status: "200",
		Kind:   "mock",
		Links:  links,
	}, nil
}

func TestStandardCrawler_FollowsLinks(t *testing.T) {
	collector := &MockCollectorWithLinks{
		Links: map[string][]string{
			"http://example.com/page1": {"http://example.com/page2"},
			"http://example.com/page2": {},
		},
	}
	registry := &MockRegistry{}

	c := crawler.NewStandardCrawler(collector, registry, 1)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() {
		c.Start(ctx, "http://example.com/page1")
	}()

	count := 0
	for range c.Results() {
		count++
		if count == 2 {
			break
		}
	}

	if count != 2 {
		t.Errorf("Expected 2 results, got %d", count)
	}
}
