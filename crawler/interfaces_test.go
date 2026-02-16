package crawler_test

import (
	"context"
	"testing"

	"github.com/jturmel/huntsman/crawler"
)

// MockCollector ensures that the Collector interface can be implemented
type MockCollector struct{}

func (m *MockCollector) Collect(ctx context.Context, targetURL string) (*crawler.Resource, error) {
	return &crawler.Resource{
		URL:        targetURL,
		Status:     "200",
		Kind:       "mock",
		Size:       100,
		Links:      []string{},
		FromSource: "",
	}, nil
}

// MockRegistry ensures that the Registry interface can be implemented
type MockRegistry struct{}

func (m *MockRegistry) Visit(u string) bool {
	return true
}

func (m *MockRegistry) IsVisited(u string) bool {
	return false
}

// MockCrawler ensures that the Crawler interface can be implemented
type MockCrawler struct {
	results chan crawler.Resource
}

func (c *MockCrawler) Start(ctx context.Context, startURL string) error {
	return nil
}

func (c *MockCrawler) Stop() {
	close(c.results)
}

func (c *MockCrawler) Results() <-chan crawler.Resource {
	return c.results
}

func TestInterfaces(t *testing.T) {
	// Verify that the mocks implement the interfaces
	var _ crawler.Collector = &MockCollector{}
	var _ crawler.Registry = &MockRegistry{}
	var _ crawler.Crawler = &MockCrawler{}
}
