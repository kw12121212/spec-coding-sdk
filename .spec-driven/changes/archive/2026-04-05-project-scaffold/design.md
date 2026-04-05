# Design: project-scaffold

## Approach

1. Create `go.mod` at repository root with module path `github.com/kw12121212/spec-coding-sdk` and Go 1.25 minimum version.
2. Create standard Go directory layout:
   - `cmd/spec-coding-sdk/` — main entry point
   - `pkg/` — public library packages (exported API)
   - `internal/` — private implementation packages
   - `api/` — protobuf definitions and generated code (gRPC later)
3. Add a minimal `main.go` that prints version info and exits — enough to verify the build works.
4. Add `Makefile` with phony targets: `build`, `test`, `lint`, `fmt`, `clean`.
5. Add `.golangci.yml` enabling a default linter set (errcheck, govet, revive, staticcheck, unused).
6. Verify `go build ./...`, `go test ./...`, and `make lint` all pass.

## Key Decisions

- **Module path**: `github.com/kw12121212/spec-coding-sdk` — chosen by user.
- **Go version**: 1.25+ — matches project spec and provides latest language features.
- **Linter**: golangci-lint with a curated default set — avoids noisy defaults while catching real issues.
- **Standard Go layout**: `cmd/`, `pkg/`, `internal/`, `api/` — conventional, familiar to Go developers, aligns with the multi-interface nature of the project (SDK, JSON-RPC, HTTP, gRPC each get their own package later).

## Alternatives Considered

- **Flat layout (no pkg/internal split)**: simpler initially but becomes messy as the project grows to 14+ milestones. Rejected in favor of standard layout.
- **Task runner alternatives (just, task)**: adds a non-Go tool dependency. Makefile is universal and sufficient for this project's build needs.
- **golangci.yml full config**: could enable all linters, but too strict for early development. Starting with a sensible subset, can be tightened later.
