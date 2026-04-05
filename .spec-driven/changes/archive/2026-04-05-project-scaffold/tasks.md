# Tasks: project-scaffold

## Implementation

- [x] Create `go.mod` with module path `github.com/kw12121212/spec-coding-sdk` and Go 1.25 minimum version
- [x] Create directory structure: `cmd/spec-coding-sdk/`, `pkg/`, `internal/`, `api/`
- [x] Add placeholder `.go` files in each package directory (with `package` declaration only) so `go build ./...` passes
- [x] Create `cmd/spec-coding-sdk/main.go` with minimal `main` function that prints version and exits
- [x] Create `Makefile` with targets: `build`, `test`, `lint`, `fmt`, `clean`
- [x] Create `.golangci.yml` with default linter set (errcheck, govet, revive, staticcheck, unused)

## Testing

- [x] Verify `go build ./...` passes
- [x] Verify `go test ./...` passes
- [x] Verify `make build`, `make test`, `make lint`, `make fmt`, `make clean` all succeed
- [x] Add a trivial unit test in `cmd/spec-coding-sdk/main_test.go` to validate test infrastructure

## Verification

- [x] All tasks above pass in a clean checkout
- [x] Directory layout matches `project-structure` spec
- [x] No lint errors from `make lint`
