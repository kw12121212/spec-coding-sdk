# Questions: tool-glob

## Open

<!-- No open questions -->

## Resolved

- [x] Q: Should the glob tool use Go's standard library or shell out to an external command?
  Context: Determines the implementation approach and external dependencies.
  A: Go standard library — no external dependency, cross-platform, consistent with Go-first approach.

- [x] Q: Should the glob tool support recursive `**` patterns?
  Context: `**` is the most common use case but Go's `filepath.Glob` does not support it natively.
  A: Yes, support `**` patterns with a custom matcher using `filepath.WalkDir`.
