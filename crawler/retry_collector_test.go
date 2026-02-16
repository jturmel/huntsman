package crawler_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jturmel/huntsman/crawler"
)

type FailingCollector struct {
	Attempts int
}

func (c *FailingCollector) Collect(ctx context.Context, url string) (*crawler.Resource, error) {
	c.Attempts++
	if c.Attempts < 3 {
		return nil, errors.New("fail")
	}
	return &crawler.Resource{URL: url, Status: "200"}, nil
}

func TestRetryCollector_Collect(t *testing.T) {
	fc := &FailingCollector{}
	rc := crawler.NewRetryCollector(fc, 3, 10*time.Millisecond)

	res, err := rc.Collect(context.Background(), "http://example.com")
	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	if fc.Attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", fc.Attempts)
	}

	if res.Status != "200" {
		t.Errorf("Expected status 200, got %s", res.Status)
	}
}
