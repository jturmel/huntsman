# Plan: Switch license to MIT

This plan outlines the steps to migrate the project license from GPLv3 to MIT.

## Phase 1: Preparation and Verification
- [ ] Task: Audit codebase for all files containing GPLv3 references.
    - [ ] Run `grep -r "General Public License" .` and `grep -r "GPL" .` to identify all files.
- [ ] Task: Conductor - User Manual Verification 'Phase 1: Preparation and Verification' (Protocol in workflow.md)

## Phase 2: License File Migration
- [ ] Task: Replace `LICENSE` file content.
    - [ ] Overwrite `LICENSE` with the standard MIT License text for "Josh Turmel".
- [ ] Task: Update `README.md`.
    - [ ] Update license section and links from GPLv3 to MIT.
- [ ] Task: Conductor - User Manual Verification 'Phase 2: License File Migration' (Protocol in workflow.md)

## Phase 3: Header Replacement
- [ ] Task: Update `main.go`.
    - [ ] Replace GPLv3 header with MIT notice.
- [ ] Task: Update `install.sh`.
    - [ ] Replace GPLv3 header with MIT notice.
- [ ] Task: Update any other files identified in Phase 1 that currently have headers.
- [ ] Task: Conductor - User Manual Verification 'Phase 3: Header Replacement' (Protocol in workflow.md)

## Phase 4: Final Validation
- [ ] Task: Verify no GPL references remain.
    - [ ] Run final grep search to ensure complete removal.
- [ ] Task: Conductor - User Manual Verification 'Phase 4: Final Validation' (Protocol in workflow.md)
