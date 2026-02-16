# Specification: Setup Dependabot with Auto-Merge

## Overview
This track involves setting up GitHub Dependabot to automatically keep the project's dependencies up to date. It includes monitoring Go modules and GitHub Actions, with an automated merging workflow for low-risk patch updates.

## Functional Requirements
1.  **Dependabot Configuration:** Create a `.github/dependabot.yml` file monitoring `gomod` and `github-actions` ecosystems.
2.  **Daily Checks:** Schedule update checks to run daily.
3.  **PR Limit:** Limit open pull requests to 5.
4.  **Auto-Merge Workflow:** Implement a GitHub Action to automatically enable "Auto-Merge" for Dependabot PRs that are classified as `patch` updates.
5.  **Safety Gate:** Merging MUST only occur after all CI status checks (tests) have passed.

## Technical Details
- Dependabot Path: `.github/dependabot.yml`
- Auto-Merge Path: `.github/workflows/dependabot-auto-merge.yml`
- Logic: Use `fetch-metadata` action to determine update type and `gh pr merge --auto` for merging.

## Acceptance Criteria
- [ ] `.github/dependabot.yml` is correctly configured and present.
- [ ] A GitHub Action is created that triggers on Dependabot PRs.
- [ ] Automated tests pass for a simulated Dependabot PR scenario.
- [ ] PRs for patch updates have "Auto-merge" enabled automatically.

## Out of Scope
- Automatic merging for minor or major version updates.
- Support for ecosystems other than Go and GitHub Actions.
