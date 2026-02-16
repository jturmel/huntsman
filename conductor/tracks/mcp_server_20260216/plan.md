# Implementation Plan: MCP Server Mode

This plan outlines the steps to implement a Model Context Protocol (MCP) server for `huntsman`, allowing AI agents to perform site crawls via a `crawl_site` tool.

## Phase 1: MCP Core & Subcommand

- [ ] Task: Research and select a Go MCP SDK (e.g., `github.com/metoro-io/mcp-golang` or `github.com/mark3labs/mcp-go`)
- [ ] Task: Create `mcp.go` to house the MCP server logic
- [ ] Task: Implement a basic MCP server with `stdio` transport
- [ ] Task: Add the `mcp` subcommand to `main.go`
- [ ] Task: Verify the server starts and responds to basic MCP lifecycle messages
- [ ] Task: Conductor - User Manual Verification 'Phase 1: MCP Core & Subcommand' (Protocol in workflow.md)

## Phase 2: `crawl_site` Tool Implementation

- [ ] Task: Register the `crawl_site` tool with required and optional parameters
- [ ] Task: Implement the bridge between the MCP tool and the existing `crawler` package
- [ ] Task: Implement logic to filter results to only include documents (HTML)
- [ ] Task: Implement the Markdown tree formatter for the tool output
- [ ] Task: Ensure `max_depth`, `max_pages`, and `headless` parameters are correctly passed to the crawler
- [ ] Task: Conductor - User Manual Verification 'Phase 2: crawl_site Tool Implementation' (Protocol in workflow.md)

## Phase 3: Documentation & Verification

- [ ] Task: Update `README.md` with a "Using as an MCP Server" section
- [ ] Task: Provide specific configuration instructions and a `.geminiignore` or config snippet for Gemini CLI
- [ ] Task: Manually verify the tool using a local MCP inspector or the Gemini CLI itself
- [ ] Task: Conductor - User Manual Verification 'Phase 3: Documentation & Verification' (Protocol in workflow.md)
