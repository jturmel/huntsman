# Conductor Protocol

This document defines the protocols followed by AI agents for managing tracks in this repository.

## 1.0 NEW TRACK INITIALIZATION (`/conductor:newTrack`)

### 1.1 Initialization & Branching
1.  **Get Track Description:** Prompt the user for a description of the track.
2.  **Infer Track Type:** Determine if it's a `feature`, `bug`, `chore`, or `refactor`.
3.  **Generate Track ID:** Create a unique ID: `shortname_YYYYMMDD`.
4.  **Create & Checkout Branch:**
    - Execute: `git checkout -b new-track/<track_id>`
    - If the branch already exists or the command fails, halt and inform the user.

### 1.2 Specification Generation
1.  **Questioning Phase:** Ask sequential questions (3-5 for features, 2-3 for others) to build the `spec.md`.
2.  **Draft `spec.md`:** Present to the user for approval.

### 1.3 Plan Generation
1.  **Generate `plan.md`:** Create a hierarchical list of tasks based on the spec and `workflow.md`.
2.  **Inject Phase Completion Tasks:** Add verification tasks at the end of each phase.
3.  **User Confirmation:** Present to the user for approval.

### 1.4 Artifact Creation
1.  **Create Directory:** `conductor/tracks/<track_id>/`
2.  **Create Metadata:** `conductor/tracks/<track_id>/metadata.json`
3.  **Write Files:** `spec.md`, `plan.md`, `index.md` to the track directory.
4.  **Update Registry:** Append the track to `conductor/tracks.md`.

### 1.5 Finalization & PR Preparation
1.  **Squash Changes:** Stage all changes and commit with `chore(conductor): Initialize track '<track_id>'`.
2.  **Push Branch:** Push the `new-track/<track_id>` branch to the remote repository.
3.  **PR Instructions:** Provide a link or instructions for the user to create a Pull Request.

## 2.0 TRACK IMPLEMENTATION (`/conductor:implement`)

1.  **Select Track:** Identify the next incomplete track or use a user-provided name.
2.  **Mark In Progress:** Update status to `[~]` in `conductor/tracks.md`.
3.  **Load Context:** Read `spec.md`, `plan.md`, and `workflow.md`.
4.  **Execute Tasks:** Follow the "Standard Task Workflow" in `workflow.md` for each task.
5.  **Finalize Track:** Update status to `[x]` in `conductor/tracks.md`.
6.  **Synchronize Docs:** Update project-level docs (Product Definition, etc.).
7.  **Cleanup:** Offer to archive or delete the track.
