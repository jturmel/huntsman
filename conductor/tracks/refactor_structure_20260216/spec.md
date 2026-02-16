# Specification: Refactor Project Structure (Internal Directory)

## Overview
This track involves reorganizing the project's Go source code to follow the standard Go project layout. By moving the core logic and UI implementation into an `internal/` directory, we clean up the root directory and clearly define the project's internal boundaries.

## Functional Requirements
1.  **Create Directory Structure:**
    - Create a `cmd/huntsman/` directory for the application entry point.
    - Create an `internal/` directory for private code.
    - Create an `internal/ui/` directory for TUI-related logic.
    - Create an `internal/utils/` directory for utilities.
2.  **Relocate Files:**
    - Move `main.go` to `cmd/huntsman/main.go`.
    - Move `ui.go`, `theme.go`, and `theme.json` to `internal/ui/`.
    - Move `utils.go` to `internal/utils/`.
    - Move the `crawler/` directory to `internal/crawler/`.
3.  **Update Imports:** Update all internal import paths to reflect the new directory structure.
4.  **Update Build System:** Update the `Makefile` and any scripts to point to the new location of `main.go`.

## Non-Functional Requirements
1.  **Standard Layout:** Adheres to the Go Project Layout conventions.
2.  **Internal Boundary:** Ensures that core logic and UI components are not accessible to external importers.

## Acceptance Criteria
- [ ] The project builds successfully with `make build`.
- [ ] The application runs correctly after the move.
- [ ] No Go source files (except for `go.mod` and potentially versioning) remain in the root directory.
- [ ] All tests pass after the reorganization.

## Out of Scope
- Adding new features or fixing existing bugs during the refactor.
- Changes to the GitHub Actions workflows beyond updating build paths.
