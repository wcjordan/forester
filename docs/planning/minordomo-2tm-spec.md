# minordomo-2tm Implementation Spec: Add a CODEOWNERS for wcjordan

## Stage 1: Create .github/CODEOWNERS

### Description
Create a `.github/CODEOWNERS` file at the root of the repository with a single rule that requires `@wcjordan` as a reviewer on every pull request. GitHub reads CODEOWNERS from `.github/CODEOWNERS` by default; the `*` glob pattern matches all files so no PR can bypass the review requirement.

### Acceptance Criteria
- `.github/CODEOWNERS` exists in the repository with the content `* @wcjordan`
- No existing files are modified
- The file follows the standard GitHub CODEOWNERS format (one rule per line, `<pattern> <owner>`)
