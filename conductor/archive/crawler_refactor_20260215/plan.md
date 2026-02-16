# Implementation Plan: Crawler Refactor & SPA Support

This plan outlines the steps to refactor the crawler engine and implement support for client-side rendered (SPA) websites.

## Phase 1: Architectural Refactoring
Focus on decoupling the existing crawler logic into a more modular structure.

- [x] Task: Define Crawler and Resource Collector interfaces. [commit: 4738bcc]
- [x] Task: Refactor existing static HTML crawler to implement the new interfaces. [commit: e78d054]
    - [ ] Write unit tests for the new modular static crawler.
    - [ ] Implement the refactored static crawler logic.
- [x] Task: Implement a centralized URL Registry to manage visited states and queuing. [commit: 38e842d]
- [x] Task: Integrate new crawler architecture into main application. [commit: 60f9c57]
- [x] Task: Conductor - User Manual Verification 'Phase 1: Architectural Refactoring' (Protocol in workflow.md) [checkpoint: 785fbb5]

## Phase 2: SPA/Dynamic Content Discovery
Introduce the capability to handle client-side rendered resources.

- [x] Task: Research and select a headless browser integration (e.g., Chromedp). [commit: a0dee48]
- [x] Task: Implement a `HeadlessCrawler` that executes JavaScript before resource extraction. [commit: d6914f1]
    - [ ] Write integration tests for the `HeadlessCrawler` using a sample SPA.
    - [ ] Implement the `HeadlessCrawler` logic.
- [x] Task: Integrate the `HeadlessCrawler` into the main application flow, allowing users to toggle between static and dynamic modes. [commit: 24743ee]
- [x] Task: Improve HeadlessCrawler performance and resource extraction (Browser reuse, Hybrid fetching, Full extraction). [commit: 84133b1]
- [x] Task: Fix HeadlessCollector resource size reporting. [commit: df3bc94]
- [x] Task: Set SPA Mode as default. [commit: af6c8bb]
- [x] Task: Conductor - User Manual Verification 'Phase 2: SPA/Dynamic Content Discovery' (Protocol in workflow.md) [checkpoint: 7a16d59]

## Phase 3: Robustness and Polish
Enhance error handling and optimize performance.

- [x] Task: Implement comprehensive error handling for timeouts, retries, and malformed content in both crawler types. [commit: 9046d32]
- [x] Task: Add concurrency controls to prevent overwhelming target servers while maintaining high performance. [commit: 9046d32]
- [x] Task: Conductor - User Manual Verification 'Phase 3: Robustness and Polish' (Protocol in workflow.md) [checkpoint: 6a87c9f]
