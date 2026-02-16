# Specification: Switch license to MIT

## Overview
This track involves migrating the project's license from GNU General Public License v3.0 (GPLv3) to the MIT License. This includes updating the main `LICENSE` file, project documentation, and existing license headers in source files.

## Functional Requirements
- **License File Update**: Replace the contents of the `LICENSE` file with the standard MIT License text.
- **Copyright Attribution**: Use "Josh Turmel" as the copyright holder in the MIT License and headers.
- **License Header Replacement**: Update every file that currently contains a GPLv3 license header to use an MIT license notice instead. Based on discovery, this includes at least:
    - `main.go`
    - `install.sh`
- **Documentation Update**: Update `README.md` to reflect the change to the MIT License and update any links pointing to the license.

## Non-Functional Requirements
- **Consistency**: Ensure the copyright year remains consistent (2026) or is updated to include the current year if appropriate.
- **Cleanliness**: Ensure no remnants of the GPLv3 license text or mentions remain in the updated files.

## Acceptance Criteria
- [ ] `LICENSE` file contains the MIT License text with "Copyright (c) 2026 Josh Turmel".
- [ ] `main.go` has the GPLv3 header replaced with a short MIT notice or reference.
- [ ] `install.sh` has the GPLv3 header replaced with a short MIT notice or reference.
- [ ] `README.md` correctly identifies the project as MIT licensed.
- [ ] A project-wide search for "GPL" or "General Public License" returns no results in source code or primary documentation.

## Out of Scope
- Adding license headers to files that do not currently have them.
- Changing the licensing of third-party dependencies.
