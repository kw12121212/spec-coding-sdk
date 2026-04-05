# Tasks: tool-grep

## Implementation

- [x] Create `internal/tools/grep/grep.go` with `Input` struct, `Tool` struct, and `New(perms)` constructor
- [x] Implement `Tool.Execute` method: parse input, validate `pattern`, build rg command with flags, execute, handle output modes and exit codes
- [x] Implement 1MB output truncation logic
- [x] Implement permission check hook via `PermissionProvider`

## Testing

- [x] Lint passes (`make lint`)
- [x] Unit tests: happy path — search with match found
- [x] Unit tests: no match found (exit code 1) returns non-error result with hint
- [x] Unit tests: input validation — empty pattern, invalid JSON
- [x] Unit tests: output modes (content, files_with_matches, count)
- [x] Unit tests: optional flags (glob, type, ignore_case, context, head_limit)
- [x] Unit tests: output truncation at 1MB
- [x] Unit tests: rg not found in PATH
- [x] Unit tests: permission denied via PermissionProvider
- [x] Unit tests: nil PermissionProvider skips check
- [x] Compile-time interface check: `var _ core.Tool = (*Tool)(nil)`

## Verification

- [x] Verify implementation matches proposal scope
- [x] `make test` passes
- [x] `make lint` passes
