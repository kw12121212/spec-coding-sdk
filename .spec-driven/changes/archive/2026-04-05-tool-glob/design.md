# Design: tool-glob

## Approach

Implement `GlobTool` as a pure Go solution using `filepath.WalkDir` for recursive directory traversal and `filepath.Match` for pattern matching. The tool follows the same structural pattern as existing tools (bash, fileops, grep):

1. `Input` struct with JSON tags for deserialization
2. `Tool` struct with optional `PermissionProvider`
3. `New(perms)` constructor
4. `Execute(ctx, input)` method validating input, checking permissions, then performing the operation

The `**` pattern is handled by splitting the pattern into a base directory and a glob suffix, then walking the base directory and matching each file path against the full pattern using `doublestar`-style logic. Since Go's `filepath.Match` does not support `**`, we implement a custom matcher that handles `**` by expanding it to any number of path segments.

For sorting, results are sorted by file modification time (most recent first), matching claw-code's behavior.

## Key Decisions

1. **Go standard library only** — no external dependencies like `fd` or `find`. This keeps the tool self-contained and cross-platform, consistent with the Go-first approach.

2. **Custom `**` matching** — Go's `filepath.Glob` does not support `**`. We implement a simple `**`-aware matcher that converts `**` segments into recursive walks, since this is the most common glob use case.

3. **Output format: one path per line** — plain file paths, one per line, sorted by modification time. This is the simplest and most universally parseable format.

4. **1MB output limit** — consistent with bash and grep tools to prevent runaway output.

5. **Default path: current working directory** — when `path` is omitted, search from the process working directory.

## Alternatives Considered

- **Shell out to `find` or `fd`** — rejected because it requires an external binary, adding a runtime dependency. The Go standard library can handle this efficiently.
- **Use a third-party glob library (e.g., `github.com/bmatcuk/doublestar`)** — rejected to avoid adding a dependency for functionality that can be implemented concisely with `filepath.WalkDir` and `filepath.Match`.
- **Use `filepath.Glob` directly** — rejected because it does not support `**` for recursive matching, which is the most common use case.
