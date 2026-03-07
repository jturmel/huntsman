package crawler

import (
	"context"
	"net/url"
)

// CrawlBounded performs a breadth-first crawl starting from startURL,
// limited to maxDepth levels and maxPages document resources.
// Only resources of kind "document" are returned.
func CrawlBounded(ctx context.Context, startURL string, collector Collector, maxDepth, maxPages int) ([]Resource, error) {
	type job struct {
		url   string
		depth int
	}

	baseURL, err := url.Parse(startURL)
	if err != nil {
		return nil, err
	}

	registry := NewInMemoryRegistry()
	registry.Visit(startURL)
	queue := []job{{url: startURL, depth: 0}}

	var results []Resource

	for len(queue) > 0 && len(results) < maxPages {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
		}

		current := queue[0]
		queue = queue[1:]

		if current.depth > maxDepth {
			continue
		}

		res, err := collector.Collect(ctx, current.url)
		if err != nil {
			continue
		}

		if res.Kind != "document" {
			continue
		}

		results = append(results, *res)

		if current.depth < maxDepth {
			for _, link := range res.Links {
				parsedLink, parseErr := url.Parse(link)
				if parseErr != nil {
					continue
				}
				if parsedLink.Host == baseURL.Host && registry.Visit(link) {
					queue = append(queue, job{url: link, depth: current.depth + 1})
				}
			}
		}
	}

	return results, nil
}
