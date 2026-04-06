# Design: permissions-hooks

## Approach

Use the existing `core.PermissionProvider` hook points already present on the tool implementations and make the tool boundary the enforcement point for this milestone.

Execution model:
1. Tool receives a request and validates input as it does today.
2. If a `PermissionProvider` is configured, the tool calls `Check(ctx, operation, resource)` before any side effect.
3. If `Check` returns nil, execution proceeds normally.
4. If `Check` returns an error representing deny or need-confirmation, the tool returns `ToolResult{IsError: true, Output: err.Error()}` and exits without spawning a process, writing a file, editing a file, or otherwise mutating state.

This change treats `DecisionNeedConfirmation` as a blocked non-interactive execution, not as an implicit allow and not as a new interactive workflow. The distinct error string remains the only signal needed at this layer; later interface milestones can decide how to translate that into callbacks, tickets, prompts, or stream events.

Testing will focus on side-effect boundaries:
- process-launching tool: prove a denied / need-confirmation bash invocation does not create the marker file it would have created if executed
- mutating file tools: prove denied / need-confirmation write/edit requests leave the filesystem unchanged
- allow-path tests continue to prove normal execution works with the same hook points

No new agent or SDK abstraction is introduced in this change. The current orchestrator spec already treats tool failures as tool observations, so permission-blocked calls can reuse that behavior without adding a separate confirmation state machine.

## Key Decisions

1. **Treat `DecisionNeedConfirmation` as blocked execution in M6.** The current execution path exposes only `Check(...) error`, so there is no safe way to continue automatically. Blocking preserves safety and keeps the semantics explicit.

2. **Keep confirmation UX out of scope.** Different interface layers will need different confirmation mechanisms. Defining one now would either leak transport concerns into the core or force later milestones into a premature API shape.

3. **Enforce and test at the tool boundary.** The risk in this milestone is unintended side effects. The most direct and observable proof is that the tool itself stops before launching a process or mutating files.

4. **Preserve existing public/internal interfaces.** `core.PermissionProvider` and the existing tool constructors already give enough structure for this milestone. Adding new interfaces now would create churn without adding necessary behavior.

## Alternatives Considered

- **Add a confirmation callback or channel now:** Rejected because it would expand M6 into SDK and transport design work that belongs to M11-M14.

- **Delay blocked-execution semantics until the SDK milestone:** Rejected because M11 already depends on M6. Shipping a public facade before the permission hook behavior is settled would increase churn.

- **Introduce a new agent-level permission broker in this change:** Rejected because the current roadmap item is about execution hooks. The existing tool-level interception points already provide the necessary safety boundary.
