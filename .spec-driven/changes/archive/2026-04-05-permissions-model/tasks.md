# Tasks: permissions-model

## Implementation

- [x] Create `internal/permission/decision.go` with `Decision` type, constants (DecisionAllow, DecisionDeny, DecisionNeedConfirmation), and `String()` method
- [x] Create `internal/permission/rule.go` with `Rule` struct (OperationPattern, ResourcePattern, Decision fields) and `Match(operation, resource) bool` method
- [x] Create `internal/permission/policy.go` with `Policy` interface (Evaluate method), `StaticPolicy` struct, and `NewStaticPolicy` constructor
- [x] Create `internal/permission/default.go` with `DefaultPolicy()` (full access, no rules) and `SafePolicy()` (restrictive, hardcoded destructive bash patterns, CWD write boundary) functions
- [x] Create `internal/permission/adapter.go` with `PolicyProvider` struct implementing `core.PermissionProvider`, `NewPolicyProvider` constructor, and compile-time interface check

## Testing

- [x] Unit tests for Decision type (String method, constant values)
- [x] Unit tests for Rule.Match (exact match, wildcard match, empty resource pattern, no match cases)
- [x] Unit tests for StaticPolicy.Evaluate (first-match wins, default decision, empty rules, multiple rules)
- [x] Unit tests for DefaultPolicy (allows all operations)
- [x] Unit tests for SafePolicy (deny destructive bash, deny write outside CWD, allow reads, allow non-destructive bash, allow grep/glob)
- [x] Unit tests for PolicyProvider.Check (allow maps to nil, deny maps to error, need_confirmation maps to error, error message format)
- [x] Lint passes (`go vet ./...`)
- [x] All tests pass (`go test ./internal/permission/...`)

## Verification

- [x] Verify `PolicyProvider` satisfies `core.PermissionProvider` via compile-time check
- [x] Verify existing tests still pass (`go test ./...`)
- [x] Verify no changes to `internal/core/interfaces.go`
