# Tasks: registry-tasks

## Implementation

- [x] Add an in-memory task registry package with task types, status constants, and registry construction
- [x] Implement create, get, update, and delete-by-ID behavior with caller-supplied IDs and minimal task fields
- [x] Enforce the allowed task status transitions and preserve task state on invalid transitions

## Testing

- [x] Add unit tests for happy path, edge cases, unknown IDs, duplicate IDs, delete semantics, and invalid transitions
- [x] Run `go test ./...`
- [x] Run `golangci-lint run ./...`

## Verification

- [x] Verify the change artifacts and implementation still match the `registry-tasks` proposal and M7 scope
