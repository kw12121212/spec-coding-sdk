# core-interfaces

## What

Define the core interfaces (`Tool`, `Agent`, `Event`, `PermissionProvider`) in `internal/core/` that establish the type foundation for all subsequent milestones.

## Why

M1 milestone notes recommend completing `core-interfaces` right after `project-scaffold`. Every downstream milestone (M2 tools, M3 agent lifecycle, M4 LLM backend, M5 permissions, etc.) depends on these interfaces being in place. Without them, no concrete implementation can compile or be tested. Defining them early also surfaces design decisions while the cost of change is lowest.

## Scope

**In scope:**

- `Tool` interface — the contract every tool (bash, file ops, grep, glob, LSP, MCP) must satisfy
- `Agent` interface — the agent lifecycle contract (session management, tool invocation loop)
- `Event` interface — structured event type for the event system
- `PermissionProvider` interface — permission check contract for tool execution
- Supporting types: `ToolResult`, `EventPayload`, error types as needed
- Tests verifying interfaces can be referenced and implemented by other packages

**Out of scope:**

- Concrete tool implementations (M2)
- Agent runtime logic (M3)
- LLM integration (M4)
- Config loading (belongs to `config-loader`)
- Event system details beyond core types (belongs to `event-system-types`)
- Public SDK re-export (M10)

## Unchanged Behavior

- `go build ./...`, `go test ./...`, `make lint` MUST continue to pass
- Existing `cmd/spec-coding-sdk/main.go` behavior MUST not change
- `pkg/sdk/`, `api/proto/` packages MUST not be modified
