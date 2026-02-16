package crawler

import (
	"context"
	"time"
)

// RetryCollector wraps a Collector and retries failed attempts
type RetryCollector struct {
	collector Collector
	retries   int
	backoff   time.Duration
}

// NewRetryCollector creates a new RetryCollector
func NewRetryCollector(collector Collector, retries int, backoff time.Duration) *RetryCollector {
	return &RetryCollector{
		collector: collector,
		retries:   retries,
		backoff:   backoff,
	}
}

// Collect attempts to collect a resource, retrying on failure
func (c *RetryCollector) Collect(ctx context.Context, targetURL string) (*Resource, error) {
	var err error
	var res *Resource
	for i := 0; i <= c.retries; i++ {
		res, err = c.collector.Collect(ctx, targetURL)
		if err == nil {
			return res, nil
		}

		// If error, check context. If cancelled, abort.
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(c.backoff * time.Duration(i+1)):
			// Retry
		}
	}
	return res, err
}
