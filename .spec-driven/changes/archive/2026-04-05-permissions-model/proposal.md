# permissions-model

## What

Implement a permission model with typed decisions (Allow/Deny/NeedConfirmation), a composable rule system, a `Policy` interface for evaluating permission requests, and a default policy covering bash and file operations. Provide an adapter bridging `Policy` to the existing `core.PermissionProvider` interface so tools can consume it without changes.

## Why

M1 defined `core.PermissionProvider` and all four tools (bash, fileops, grep, glob) already call `Check(ctx, operation, resource)` — but no real implementation exists, only test mocks. This change fills that gap by providing a concrete permission model that:
- Supports three decision outcomes (allow, deny, need-confirmation) needed by both SDK embedding (auto-allow) and service scenarios (require confirmation)
- Gives downstream milestones (M7 registries, M11 SDK, M12–M14 interface layers) a working policy to build on
- Establishes the rule/policy vocabulary that `permissions-hooks` (the second M6 change) will wire into the orchestrator

## Scope

In scope:
- `Decision` type with Allow/Deny/NeedConfirmation constants
- `Rule` type: pattern-based matcher for operation and resource
- `Policy` interface: `Evaluate(ctx, operation, resource) -> Decision`
- `StaticPolicy`: ordered list of rules with a default decision
- `DefaultPolicy()`: full-access policy (allow everything, no rules)
- `SafePolicy()`: pre-built restrictive policy covering bash execution and file read/write/edit safety
- `PolicyProvider` adapter: wraps a `Policy` to satisfy `core.PermissionProvider`
- Unit tests for all types

Out of scope:
- Destructive bash pattern list is hardcoded in SafePolicy (user decision: hardcoded)
- Write boundary uses CWD captured at SafePolicy() creation time (user decision: CWD)
- DefaultPolicy is full access; SafePolicy provides restrictive defaults (user decision: full access default)
- Interface-layer-specific policy customization (M11–M14)
- Human-in-the-loop confirmation flow (future, depends on interface layer)

## Unchanged Behavior

- Existing `core.PermissionProvider` interface must not change
- Existing tool constructors (`bash.New`, `NewReadTool`, etc.) must not change
- Existing tests must continue to pass
