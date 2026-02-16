package crawler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jturmel/huntsman/crawler"
)

func TestStaticCollector_Collect(t *testing.T) {
	// Setup test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
			<html>
				<body>
					<a href="/page2">Link 1</a>
					<img src="/image.png" />
				</body>
			</html>
		`))
	}))
	defer ts.Close()

	c := crawler.NewStaticCollector()
	resource, err := c.Collect(context.Background(), ts.URL)
	if err != nil {
		t.Fatalf("Collect failed: %v", err)
	}

	if resource.Status != "200" {
		t.Errorf("Expected status 200, got %s", resource.Status)
	}

	if resource.Kind != "document" {
		t.Errorf("Expected kind document, got %s", resource.Kind)
	}

	// Expecting 2 links: /page2 and /image.png
	if len(resource.Links) != 2 {
		t.Errorf("Expected 2 links, got %d", len(resource.Links))
	}
}
