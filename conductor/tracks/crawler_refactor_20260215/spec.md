# Specification: Crawler Refactor & SPA Support

## Overview
This track focuses on refactoring the existing `crawler.go` to improve modularity and robustness, while introducing a new capability to discover resources in Client-Side Rendered (SPA) applications.

## Problem Statement
The current crawler primarily handles static HTML response parsing. As many modern websites rely on client-side JavaScript to render content and links, Huntsman currently misses these resources. Additionally, the crawling logic is tightly coupled, making it difficult to test and extend.

## Goals
- **Decouple Crawler Logic:** Separate network fetching, HTML parsing, and URL management into distinct, testable components.
- **Implement SPA Support:** Integrate a mechanism (e.g., using a headless browser or specialized library) to execute JavaScript and capture dynamically rendered links.
- **Improve Robustness:** Enhance error handling for various network conditions and malformed HTML.

## Proposed Architecture
- **Crawler Interface:** Define a clear interface for the crawler to allow for different implementations (Static vs. Headless).
- **Registry/Queue:** A centralized component to manage the crawl state and prevent redundant visits.
- **Resource Collector:** A component responsible for extracting resources from the rendered DOM.
