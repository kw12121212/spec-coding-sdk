# tool-glob

## What

Implement a file pattern matching tool (GlobTool) that finds files by glob pattern, supporting recursive `**` patterns, using Go's standard library. The tool implements the `core.Tool` interface with permission check hook support.

## Why

M2 Tool Surface requires four core tools: bash, file ops, grep, and glob. The first three are already archived. `tool-glob` is the last remaining change to complete M2, which unblocks M3 and all downstream milestones.

## Scope

- New `internal/tools/glob/` package with `Tool` struct implementing `core.Tool`
- JSON input schema with `pattern` (required) and `path` (optional) fields
- Recursive `**` pattern support using `filepath.WalkDir`
- Permission check hook via `PermissionProvider` injection
- Output limit of 1MB with truncation
- Sorted results by modification time (most recent first)
- Unit tests covering happy path, error cases, and edge cases
- Delta spec update to `tool-surface.md`

## Unchanged Behavior

- Existing tools (bash, fileops, grep) are not modified
- `core.Tool` and `core.PermissionProvider` interfaces remain unchanged
- No changes to configuration loading or event system
