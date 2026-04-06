# Tasks: permissions-hooks

## Implementation

- [x] Review the existing permission and tool implementations to identify any missing pre-execution interception or inconsistent blocked-result handling
- [x] Implement or refine blocked-execution behavior so deny and need-confirmation results stop tool execution before side effects occur
- [x] Add or update tests for the shipped tools to verify blocked executions preserve distinct error messages and do not create side effects

## Testing

- [x] Run `go vet ./...`
- [x] Run targeted unit tests covering permission allow / deny / need-confirmation paths for bash and file-mutating tools
- [x] Run `go test ./...`

## Verification

- [x] Verify the implementation matches the narrowed scope: non-interactive blocking only, no confirmation UX
- [x] Verify the observed behavior still matches the existing tool and agent contracts
- [x] Verify roadmap item `permissions-hooks` is fully covered by the resulting delta spec and tests
