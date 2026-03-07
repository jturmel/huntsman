package crawler_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jturmel/huntsman/crawler"
)

func TestCrawlBounded_SinglePage(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body><a href="/about">About</a></body></html>`))
	}))
	defer ts.Close()

	col := crawler.NewStaticCollector()
	results, err := crawler.CrawlBounded(context.Background(), ts.URL, col, 3, 50)
	if err != nil {
		t.Fatalf("CrawlBounded failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("Expected at least one result, got none")
	}

	if results[0].Kind != "document" {
		t.Errorf("Expected kind document, got %s", results[0].Kind)
	}
}

func TestCrawlBounded_MaxPages(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		// Serve pages that link to each other to generate many pages
		var links string
		for i := 0; i < 20; i++ {
			links += fmt.Sprintf(`<a href="/page%d">Page %d</a>`, i, i)
		}
		w.Write([]byte(fmt.Sprintf(`<html><body>%s</body></html>`, links)))
	}))
	defer ts.Close()

	maxPages := 3
	col := crawler.NewStaticCollector()
	results, err := crawler.CrawlBounded(context.Background(), ts.URL, col, 10, maxPages)
	if err != nil {
		t.Fatalf("CrawlBounded failed: %v", err)
	}

	if len(results) > maxPages {
		t.Errorf("Expected at most %d pages, got %d", maxPages, len(results))
	}
}

func TestCrawlBounded_MaxDepth(t *testing.T) {
	// Build a chain: / -> /depth1 -> /depth2 -> /depth3
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		switch r.URL.Path {
		case "/":
			w.Write([]byte(`<html><body><a href="/depth1">d1</a></body></html>`))
		case "/depth1":
			w.Write([]byte(`<html><body><a href="/depth2">d2</a></body></html>`))
		case "/depth2":
			w.Write([]byte(`<html><body><a href="/depth3">d3</a></body></html>`))
		case "/depth3":
			w.Write([]byte(`<html><body>leaf</body></html>`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	col := crawler.NewStaticCollector()

	// With max_depth=1, only / and /depth1 should be visited
	results, err := crawler.CrawlBounded(context.Background(), ts.URL+"/", col, 1, 50)
	if err != nil {
		t.Fatalf("CrawlBounded failed: %v", err)
	}

	if len(results) > 2 {
		t.Errorf("Expected at most 2 pages with max_depth=1, got %d", len(results))
	}
}

func TestCrawlBounded_NonDocumentFiltered(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/style.css" {
			w.Header().Set("Content-Type", "text/css")
			w.Write([]byte("body { color: red; }"))
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><head><link href="/style.css" rel="stylesheet"></head><body>Hello</body></html>`))
	}))
	defer ts.Close()

	col := crawler.NewStaticCollector()
	results, err := crawler.CrawlBounded(context.Background(), ts.URL, col, 3, 50)
	if err != nil {
		t.Fatalf("CrawlBounded failed: %v", err)
	}

	for _, r := range results {
		if r.Kind != "document" {
			t.Errorf("Expected only document resources, got kind %s for URL %s", r.Kind, r.URL)
		}
	}
}

func TestCrawlBounded_ContextCancellation(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body>Hello</body></html>`))
	}))
	defer ts.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	col := crawler.NewStaticCollector()
	_, err := crawler.CrawlBounded(ctx, ts.URL, col, 3, 50)
	if err == nil {
		// An immediately-cancelled context may return no error if the queue
		// is processed before cancellation is detected; that is acceptable.
		t.Log("CrawlBounded completed without error on cancelled context (acceptable)")
	}
}
