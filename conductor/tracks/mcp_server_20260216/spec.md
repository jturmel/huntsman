# Specification: MCP Server Mode for Huntsman

## Overview
Implement a Model Context Protocol (MCP) server mode for `huntsman`. This will allow AI agents (like Claude Desktop or Gemini CLI) to use `huntsman` as a tool to explore and map website structures directly. The focus is on providing a clean, hierarchical view of a site's pages to facilitate further agent actions.

## Functional Requirements
- **New Subcommand:** Implement `huntsman mcp` to start the MCP server.
- **Transport Protocol:** Use the **stdio** transport protocol for communication.
- **Core Tool: `crawl_site`**
    - **Inputs:**
        - `url` (string, required): The starting URL for the crawl.
        - `max_depth` (integer, optional): Maximum depth to crawl (default: 3).
        - `max_pages` (integer, optional): Maximum number of pages to crawl (default: 50).
        - `headless` (boolean, optional): Whether to use a headless browser for JavaScript rendering (default: false).
    - **Behavior:**
        - Specifically target document/page resources (HTML).
        - Filter out static assets like images, CSS, and scripts to keep the output focused on site structure.
    - **Output:**
        - A structured Markdown tree representation of the discovered URLs (e.g., nested bullet points).
        - This format is optimized for LLM consumption to help the agent understand the hierarchy and decide on next steps.
- **Documentation:**
    - Update `README.md` with instructions on how to configure and use the `huntsman` MCP server with **Gemini CLI**.

## Non-Functional Requirements
- **Modular Design:** The MCP server implementation should be decoupled from the transport layer to allow for future SSE (HTTP) support.
- **Efficiency:** The crawl should honor the provided limits (`max_depth`, `max_pages`) strictly to avoid excessive resource usage or context window bloat.

## Acceptance Criteria
- Running `huntsman mcp` starts a valid MCP server responding on stdin/stdout.
- The `crawl_site` tool is correctly exposed to MCP clients.
- `crawl_site` successfully crawls a given URL up to the specified depth/page limits.
- The tool returns a Markdown tree of page URLs.
- Headless mode works as expected when toggled via the tool parameter.
- `README.md` contains clear configuration examples for Gemini CLI.

## Out of Scope
- Support for authenticated crawls (cookies/headers).
- Advanced rate-limiting configuration via tool parameters.
- SSE (HTTP) transport protocol (reserved for future work).
- Crawling non-document assets (images, PDFs, etc.) unless requested in a later track.
