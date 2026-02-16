package crawler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jturmel/huntsman/crawler"
)

func TestHeadlessCollector_Collect(t *testing.T) {
	// Setup test server with JS
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
			<html>
				<body>
					<div id="content"></div>
					<script>
						setTimeout(() => {
							var a = document.createElement('a');
							a.href = "/dynamic";
							a.innerText = "Dynamic Link";
							document.getElementById("content").appendChild(a);
						}, 100);
					</script>
				</body>
			</html>
		`))
	}))
	defer ts.Close()

	c := crawler.NewHeadlessCollector()

	// Collect
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	resource, err := c.Collect(ctx, ts.URL)
	if err != nil {
		t.Fatalf("Collect failed: %v", err)
	}

	found := false
	expectedLink := ts.URL + "/dynamic"
	
	for _, link := range resource.Links {
		if link == expectedLink {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected to find dynamic link, got %v", resource.Links)
	}

	if resource.Size <= 0 {
		t.Errorf("Expected resource size > 0, got %d", resource.Size)
	}
}
