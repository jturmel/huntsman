# Specification: Auto-branching for New Tracks

## Overview
This track modifies the `conductor:newTrack` workflow to automatically create and checkout a new Git branch before any track artifacts (specification, implementation plan, metadata) are created. This ensures that all documentation related to the track's inception is isolated.

## Functional Requirements
1.  **Early Branching:** The branch creation and checkout MUST occur at the beginning of the `conductor:newTrack` process, specifically before any files are written to the disk.
2.  **Naming Convention:** The initial branch MUST follow the naming convention `new-track/<track_id>`.
    *   `<track_id>`: The generated unique ID for the track (e.g., `shortname_YYYYMMDD`).
    *   Example: `new-track/auto_branching_20260215`.
3.  **Base Branch:** The new branch MUST be branched from the current active branch (typically `main` or `master`).
4.  **Error Handling:** If the branch already exists or if the checkout fails, the process MUST halt and inform the user.

## Non-Functional Requirements
1.  **Workflow Consistency:** The change must integrate seamlessly with existing Conductor protocols and maintain the integrity of the Tracks Registry.

## Acceptance Criteria
- [ ] Running `/conductor:newTrack` successfully creates a new Git branch named `new-track/<track_id>`.
- [ ] The new branch is checked out before `spec.md`, `plan.md`, and `metadata.json` are created.
- [ ] The track files are present on the new branch and NOT on the original branch.
- [ ] The Tracks Registry update and commit happen on the new branch.

## Out of Scope
- Automatic creation of the `feat/`, `fix/`, etc., branches (this remains part of the implementation start).
- Automatic merging of the branch.
