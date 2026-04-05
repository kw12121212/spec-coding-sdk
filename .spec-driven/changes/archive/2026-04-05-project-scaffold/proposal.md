# project-scaffold

## What

Initialize the Go module, project directory structure, Makefile, and lint configuration. This is the foundational change that creates the build infrastructure all subsequent changes depend on.

## Why

M1 is the first milestone and `project-scaffold` is the entry point. Without a Go module, directory layout, and build tooling, no other change can compile or run tests. This change must land first.

## Scope

**In scope:**

- `go.mod` with module path `github.com/kw12121212/spec-coding-sdk`, Go 1.25+
- Standard Go directory layout: `cmd/`, `pkg/`, `internal/`, `api/`
- `Makefile` with targets: `build`, `test`, `lint`, `fmt`, `clean`
- `.golangci.yml` with a sensible default linter set
- Minimal `main.go` under `cmd/` that compiles and runs
- `go build ./...` and `go test ./...` pass from the start

**Out of scope:**

- Core interface definitions (belongs to `core-interfaces`)
- Config loading (belongs to `config-loader`)
- Event types (belongs to `event-system-types`)
- Any business logic or functional code

## Unchanged Behavior

N/A — this is the first change; no existing behavior to preserve.
