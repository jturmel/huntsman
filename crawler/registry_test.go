package crawler_test

import (
	"sync"
	"testing"

	"github.com/jturmel/huntsman/crawler"
)

func TestInMemoryRegistry_Visit(t *testing.T) {
	registry := crawler.NewInMemoryRegistry()

	// First visit should return true
	if !registry.Visit("http://example.com") {
		t.Error("First visit should return true")
	}

	// Second visit should return false
	if registry.Visit("http://example.com") {
		t.Error("Second visit should return false")
	}

	// Different URL should return true
	if !registry.Visit("http://example.com/page2") {
		t.Error("Visit to new URL should return true")
	}
}

func TestInMemoryRegistry_IsVisited(t *testing.T) {
	registry := crawler.NewInMemoryRegistry()

	registry.Visit("http://example.com")

	if !registry.IsVisited("http://example.com") {
		t.Error("Expected IsVisited to return true")
	}

	if registry.IsVisited("http://example.org") {
		t.Error("Expected IsVisited to return false")
	}
}

func TestInMemoryRegistry_Concurrency(t *testing.T) {
	registry := crawler.NewInMemoryRegistry()
	var wg sync.WaitGroup
	count := 0
	var mu sync.Mutex

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if registry.Visit("http://example.com") {
				mu.Lock()
				count++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	if count != 1 {
		t.Errorf("Expected exactly 1 successful visit, got %d", count)
	}
}
