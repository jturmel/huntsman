# Implementation Plan: Refactor Project Structure (Internal Directory)

This plan outlines the steps to reorganize the project structure into `cmd/` and `internal/` directories.

## Phase 1: Directory Setup & File Relocation

- [ ] Task: Create new directory structure (`cmd/huntsman/`, `internal/ui/`, `internal/utils/`, `internal/crawler/`)
- [ ] Task: Move `main.go` to `cmd/huntsman/`
- [ ] Task: Move `ui.go`, `theme.go`, and `theme.json` to `internal/ui/`
- [ ] Task: Move `utils.go` to `internal/utils/`
- [ ] Task: Move `crawler/` contents to `internal/crawler/`
- [ ] Task: Conductor - User Manual Verification 'Phase 1: Directory Setup & File Relocation' (Protocol in workflow.md)

## Phase 2: Code & Build System Updates

- [ ] Task: Update package names and internal imports in all Go files
- [ ] Task: Update `Makefile` to reflect the new `main.go` path
- [ ] Task: Update `install.sh` if it references specific source paths
- [ ] Task: Conductor - User Manual Verification 'Phase 2: Code & Build System Updates' (Protocol in workflow.md)

## Phase 3: Verification

- [ ] Task: Verify the project builds with `make build`
- [ ] Task: Run all tests with `go test ./...`
- [ ] Task: Verify the application runs correctly
- [ ] Task: Conductor - User Manual Verification 'Phase 3: Verification' (Protocol in workflow.md)
