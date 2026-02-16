# Specification: Modernized Install & Release Process

## Overview
This track refactors the installation and release processes to align with modern best practices, drawing inspiration from the `cleat` project. This includes a new `install.sh` script and a GitHub Actions workflow for automated releases.

## Functional Requirements
1.  **Compressed Asset Support (Installer):**
    - Transition `install.sh` to download and extract `.tar.gz` archives.
2.  **Intelligent Permission Handling (Installer):**
    - Conditionally use `sudo` only when write permissions are missing for the target directory.
3.  **Refined Installation Scope:**
    - Focus exclusively on the binary installation, removing `theme.json` handling.
4.  **Automated Release Workflow (GitHub Actions):**
    - Create/Update a workflow that triggers automatically when a tag matching `v*.*.*` is pushed.
    - The workflow MUST build binaries for multiple platforms (Linux/macOS) and architectures (amd64/arm64).
    - The workflow MUST package these binaries into `.tar.gz` archives and attach them to a new GitHub Release.

## Technical Details
- **Installer:** Refactor `install.sh` to handle archive extraction and intelligent sudo.
- **Workflow:** Implement `.github/workflows/release.yml` (or similar) using standard Go build and release actions.

## Acceptance Criteria
- [ ] `install.sh` successfully installs `huntsman` from a `.tar.gz` asset.
- [ ] Pushing a tag like `v1.0.0` triggers a GitHub Action.
- [ ] The Action successfully builds, packages, and releases assets for Linux/macOS (amd64/arm64).
- [ ] The release assets match the expected naming convention (e.g., `huntsman_v1.0.0_linux_amd64.tar.gz`).

## Out of Scope
- Implementing the application's internal configuration initialization.
