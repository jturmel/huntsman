package crawler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// StaticCollector implements the Collector interface for static HTML pages
type StaticCollector struct {
	client *http.Client
}

// NewStaticCollector creates a new StaticCollector with a default HTTP client
func NewStaticCollector() *StaticCollector {
	return &StaticCollector{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Collect fetches the targetURL and extracts resources
func (c *StaticCollector) Collect(ctx context.Context, targetURL string) (*Resource, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		// Return resource with Error status to indicate failure but preserve URL
		return &Resource{URL: targetURL, Status: "Error", Kind: "N/A"}, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return &Resource{URL: targetURL, Status: "Read Err", Kind: "N/A"}, err
	}

	contentType := resp.Header.Get("Content-Type")
	kind := DetermineKind(contentType)

	var links []string
	if kind == "document" {
		links = extractLinks(strings.NewReader(string(bodyBytes)), targetURL)
	}

	return &Resource{
		URL:        targetURL,
		Status:     fmt.Sprintf("%d", resp.StatusCode),
		Kind:       kind,
		Size:       int64(len(bodyBytes)),
		Links:      links,
		FromSource: "", // Caller manages source attribution
	}, nil
}


func extractLinks(body io.Reader, currentUrl string) []string {
	var links []string
	z := html.NewTokenizer(body)

	baseUrl, err := url.Parse(currentUrl)
	if err != nil {
		return links
	}

	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return links
		case html.StartTagToken, html.SelfClosingTagToken:
			t := z.Token()
			var attrKey string
			switch t.Data {
			case "a", "link":
				attrKey = "href"
			case "img", "script", "video", "audio", "source":
				attrKey = "src"
			default:
				continue
			}

			for _, a := range t.Attr {
				if a.Key == attrKey {
					u, err := url.Parse(a.Val)
					if err != nil {
						continue
					}
					resolved := baseUrl.ResolveReference(u)
					// Note: Removed same-domain check here to allow Collector to be pure.
					// Filtering should happen in the Crawler/Registry.
					resolved.Fragment = ""
					links = append(links, resolved.String())
				}
			}
		}
	}
}
