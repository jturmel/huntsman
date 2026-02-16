package crawler

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

// HeadlessCollector uses a headless browser to collect resources
type HeadlessCollector struct {
}

// NewHeadlessCollector creates a new HeadlessCollector
func NewHeadlessCollector() *HeadlessCollector {
	return &HeadlessCollector{}
}

// Collect navigates to the URL and extracts links from the rendered DOM
func (c *HeadlessCollector) Collect(ctx context.Context, targetURL string) (*Resource, error) {
	var size int64
	var status string = "200" // Default

	// Hybrid Check: Use HEAD request first
	// We use a short timeout for the HEAD request to fail fast if it's not available
	headCtx, cancelHead := context.WithTimeout(ctx, 5*time.Second)
	req, err := http.NewRequestWithContext(headCtx, "HEAD", targetURL, nil)
	if err == nil {
		resp, err := http.DefaultClient.Do(req)
		cancelHead() // Cancel HEAD context immediately after response
		if err == nil {
			defer resp.Body.Close()
			size = resp.ContentLength
			status = fmt.Sprintf("%d", resp.StatusCode)
			
			ctype := resp.Header.Get("Content-Type")
			kind := DetermineKind(ctype)
			// If it's a known non-document type, return immediately as static resource
			if kind != "document" && kind != "Other" {
				return &Resource{
					URL:        targetURL,
					Status:     status,
					Kind:       kind,
					Size:       size,
					Links:      []string{},
					FromSource: "",
				}, nil
			}
		}
	} else {
		cancelHead()
	}

	// Proceed with browser navigation
	// Create new tab context from passed context (reusing browser if present)
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	// Ensure timeout for navigation and extraction
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var nodes []*cdp.Node
	res := &Resource{
		URL:    targetURL,
		Kind:   "document", // Assume document if we are here
		Status: status,     // Use status from HEAD check if available
		Size:   size,       // Use size from HEAD check
	}

	// Selector for all resources we care about
	selector := "a[href], link[href], img[src], script[src]"

	// Run tasks
	err = chromedp.Run(ctx,
		chromedp.Navigate(targetURL),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		// Wait a bit for JS to execute (simple heuristic)
		chromedp.Sleep(2*time.Second),
		chromedp.Nodes(selector, &nodes, chromedp.ByQueryAll),
	)

	if err != nil {
		res.Status = "Error"
		return res, err
	}

	var links []string
	baseURL, _ := url.Parse(targetURL)

	for _, n := range nodes {
		var rawURL string
		nodeName := strings.ToUpper(n.NodeName)
		if nodeName == "A" || nodeName == "LINK" {
			rawURL = n.AttributeValue("href")
		} else if nodeName == "IMG" || nodeName == "SCRIPT" {
			rawURL = n.AttributeValue("src")
		}

		if rawURL != "" {
			u, err := url.Parse(rawURL)
			if err != nil {
				continue
			}
			resolved := baseURL.ResolveReference(u)
			resolved.Fragment = ""
			links = append(links, resolved.String())
		}
	}
	res.Links = links

	return res, nil
}
