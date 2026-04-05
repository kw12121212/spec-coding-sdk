# Tasks: tool-glob

## Implementation

- [x] Create `internal/tools/glob/glob.go` with `Input` struct, `Tool` struct, `New()` constructor, and `Execute()` method
- [x] Implement `**`-aware pattern matching via `filepath.WalkDir` + custom matcher
- [x] Implement permission check hook (`Check(ctx, "glob:execute", pattern)`)
- [x] Implement 1MB output truncation
- [x] Sort results by modification time (most recent first)

## Testing

- [x] Unit test: basic glob pattern matches files in a directory
- [x] Unit test: `**` recursive pattern matches files in nested directories
- [x] Unit test: empty pattern returns error
- [x] Unit test: non-existent path returns error result
- [x] Unit test: permission check blocks execution when denied
- [x] Unit test: permission check skipped when provider is nil
- [x] Unit test: output truncation at 1MB limit
- [x] `go vet ./internal/tools/glob/` passes
- [x] `go test ./internal/tools/glob/` passes

## Verification

- [x] Verify implementation matches proposal scope
- [x] Verify delta spec reflects actual behavior
- [x] Verify `go test ./...` still passes globally
