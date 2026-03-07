package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/jturmel/huntsman/crawler"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// runMCPServer starts an MCP server over stdio, exposing the crawl_site tool.
func runMCPServer() error {
	s := server.NewMCPServer(
		"huntsman",
		version,
		server.WithToolCapabilities(false),
	)

	crawlTool := mcp.NewTool("crawl_site",
		mcp.WithDescription("Crawl a website and return a Markdown list of all discovered pages within the domain."),
		mcp.WithString("url",
			mcp.Required(),
			mcp.Description("The starting URL to crawl (e.g. https://example.com)."),
		),
		mcp.WithNumber("max_depth",
			mcp.Description("Maximum crawl depth from the starting URL (default: 3)."),
		),
		mcp.WithNumber("max_pages",
			mcp.Description("Maximum number of pages to collect (default: 50)."),
		),
		mcp.WithBoolean("headless",
			mcp.Description("Use a headless browser for JavaScript-rendered pages (default: false)."),
		),
	)

	s.AddTool(crawlTool, crawlSiteHandler)

	return server.ServeStdio(s)
}

func crawlSiteHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	targetURL, err := req.RequireString("url")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	maxDepth := req.GetInt("max_depth", 3)
	maxPages := req.GetInt("max_pages", 50)
	headless := req.GetBool("headless", false)

	var col crawler.Collector
	if headless {
		col = crawler.NewHeadlessCollector()
	} else {
		col = crawler.NewStaticCollector()
	}

	resources, crawlErr := crawler.CrawlBounded(ctx, targetURL, col, maxDepth, maxPages)
	if crawlErr != nil {
		return mcp.NewToolResultError(fmt.Sprintf("crawl failed: %v", crawlErr)), nil
	}

	return mcp.NewToolResultText(buildMarkdownTree(resources, targetURL)), nil
}

func buildMarkdownTree(resources []crawler.Resource, startURL string) string {
	if len(resources) == 0 {
		return fmt.Sprintf("No pages found at %s.", startURL)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Site Map: %s\n\n", startURL))
	for _, r := range resources {
		sb.WriteString(fmt.Sprintf("- %s\n", r.URL))
	}
	return sb.String()
}
