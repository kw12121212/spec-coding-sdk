# Design: permissions-model

## Approach

Create a new `internal/permission/` package containing the permission model types, policy interface, and default implementation. Provide a `PolicyProvider` adapter struct that implements `core.PermissionProvider` by delegating to a `Policy`.

Package layout:
```
internal/permission/
  decision.go     — Decision type and constants
  rule.go         — Rule struct and matching logic
  policy.go       — Policy interface and StaticPolicy implementation
  default.go      — DefaultPolicy() constructor
  adapter.go      — PolicyProvider adapter (Policy → core.PermissionProvider)
```

Flow: tool calls `PermissionProvider.Check(ctx, op, res)` → adapter delegates to `Policy.Evaluate(ctx, op, res)` → Policy evaluates rules in order → returns Decision → adapter maps Allow→nil, Deny→error, NeedConfirmation→error.

## Key Decisions

1. **New package `internal/permission/`** rather than extending `internal/core/`. The permission model is a distinct subsystem. Keeping it separate avoids bloating the core package and makes it easier to test in isolation.

2. **`Decision` as a string-based type** (like `agent.State`). Provides human-readable string representation, JSON-serializable, and extensible without enum boilerplate.

3. **`StaticPolicy` as ordered rule list with a default decision.** Simple, predictable, and covers the current use case. Each rule has an operation pattern and optional resource pattern (both using `filepath.Match` syntax). First matching rule wins; if no rule matches, the default decision applies.

4. **Adapter maps NeedConfirmation → Deny at the `PermissionProvider` level.** The current `Check()` signature can only express allow (nil) or deny (error). NeedConfirmation requires an interactive channel that doesn't exist yet. The adapter treats it as a deny with a distinct error message, so callers can distinguish it if needed. The orchestrator hook (`permissions-hooks`) will handle the confirmation flow later.

5. **DefaultPolicy is full access; SafePolicy provides restrictive defaults.** `DefaultPolicy()` returns an allow-everything policy for SDK embedding where the consumer manages their own security. `SafePolicy()` provides the restrictive baseline (deny destructive bash, deny writes outside CWD) for service scenarios. This matches the user's requirement that the default be permissive.

## Alternatives Considered

- **Extend `core.PermissionProvider` to return `(Decision, error)`:** Breaking change to all tool implementations. Rejected to maintain backward compatibility.

- **Map-based policy (operation → Decision):** Simpler but cannot express resource-dependent rules (e.g., "allow file:read everywhere, deny file:write outside CWD"). Rejected in favor of pattern-based rules.

- **Separate `NeedConfirmation` handling via callback in adapter:** Premature — the confirmation flow depends on interface-layer specifics (M11–M14). Deferred to `permissions-hooks`.

- **Embed permission types in `internal/core/`:** Would couple the permission model to core interfaces. Rejected for separation of concerns.
