# Tasks: config-loader

## Implementation

- [x] Add `gopkg.in/yaml.v3` dependency via `go get`
- [x] Create `internal/core/config.go` with `Config` struct and `LoadConfig(path string) (*Config, error)` function
- [x] Implement file-not-found and YAML parse error handling with descriptive messages

## Testing

- [x] Create `internal/core/config_test.go` with tests:
  - Load a valid (empty) YAML file returns `*Config` with no error
  - Load a non-existent file returns error containing the file path
  - Load a malformed YAML file returns error describing the parse failure
- [x] Run `go test ./...` — all tests pass
- [x] Run `make lint` — no new warnings

## Verification

- [x] `go build ./...` passes
- [x] `Config` struct is importable from other packages (verified by test in external test package)
