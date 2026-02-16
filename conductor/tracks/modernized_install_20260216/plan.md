# Implementation Plan: Modernized Install & Release Process

This plan outlines the steps to refactor the installation script and implement an automated release workflow based on the `cleat` project's model.

## Phase 1: Automated Release Workflow

- [ ] Task: Create or update `.github/workflows/release.yml`
- [ ] Task: Implement build matrix for Linux/macOS (amd64/arm64)
- [ ] Task: Implement packaging logic to create `.tar.gz` archives
- [ ] Task: Configure the workflow to trigger on `v*.*.*` tags and create a GitHub Release
- [ ] Task: Conductor - User Manual Verification 'Phase 1: Automated Release Workflow' (Protocol in workflow.md)

## Phase 2: Refactor Installation Script

- [ ] Task: Update `install.sh` to handle `.tar.gz` assets
- [ ] Task: Implement intelligent `sudo` logic for directory permissions
- [ ] Task: Remove `theme.json` installation logic from `install.sh`
- [ ] Task: Update version detection to use the GitHub API or similar robust method
- [ ] Task: Conductor - User Manual Verification 'Phase 2: Refactor Installation Script' (Protocol in workflow.md)

## Phase 3: Final Integration & Verification

- [ ] Task: Verify the full flow by creating a test tag (if possible in a fork/local simulation)
- [ ] Task: Verify the installer works with the newly structured release assets
- [ ] Task: Conductor - User Manual Verification 'Phase 3: Final Integration & Verification' (Protocol in workflow.md)
