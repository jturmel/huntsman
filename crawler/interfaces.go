package crawler

import (
	"context"
)

// Resource represents a discovered resource (URL, script, image, etc.)
type Resource struct {
	URL        string
	Status     string // Use string to support "Error" states
	Kind       string // e.g., "document", "script", "image"
	Size       int64
	Links      []string // Outgoing links found on this resource
	FromSource string   // The referrer URL where this resource was found
}

// Collector is responsible for fetching and parsing a single resource
type Collector interface {
	Collect(ctx context.Context, targetURL string) (*Resource, error)
}

// Crawler manages the crawling process (concurrency, queue, results)
type Crawler interface {
	Start(ctx context.Context, startURL string) error
	Stop()
	Results() <-chan Resource
}

// Registry manages the visited state of URLs to prevent redundant processing
type Registry interface {
	Visit(u string) bool // Returns true if the URL was visited for the first time
	IsVisited(u string) bool
}
