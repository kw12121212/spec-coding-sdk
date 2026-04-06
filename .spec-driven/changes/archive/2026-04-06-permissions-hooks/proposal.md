# permissions-hooks

## What

Complete the second M6 planned change by defining and implementing the remaining permission-hook execution behavior for non-interactive tool calls. This change makes denied and need-confirmation decisions block execution before any tool side effects occur, surfaces stable error results back to callers, and adds tests proving the shipped tools honor that boundary.

## Why

The repository already has the permission model (`Decision`, `Policy`, `PolicyProvider`) and the core tools expose `PermissionProvider` hook points, but the roadmap item `permissions-hooks` is still open. Before the project exposes a public SDK or remote interfaces, the permission system needs one clear contract for what happens when a tool call is denied or requires confirmation in an execution path that has no human confirmation channel.

Keeping this change narrow avoids prematurely locking in SDK / JSON-RPC / HTTP / gRPC confirmation UX while still finishing M6's safety baseline.

## Scope

In scope:
- Clarify the observable behavior of `DecisionNeedConfirmation` in non-interactive execution paths
- Finish and verify permission interception behavior for the currently shipped tools using existing `PermissionProvider` hook points
- Prove with tests that blocked executions do not spawn processes or mutate files
- Keep the returned error messages distinct so callers can tell deny from need-confirmation

Out of scope:
- Human approval UX, callbacks, pause/resume flows, or interactive confirmation sessions
- SDK / JSON-RPC / HTTP / gRPC specific permission APIs
- New permission model types or policy backends beyond the existing `internal/permission/` package
- Registry, MCP, LSP, or other unrelated roadmap items

## Unchanged Behavior

Behaviors that must not change as a result of this change (leave blank if nothing is at risk):
- `core.PermissionProvider` remains the permission hook contract
- Existing tool constructors (`bash.New`, `fileops.NewReadTool`, `fileops.NewWriteTool`, `fileops.NewEditTool`, `grep.New`, `glob.New`) remain unchanged
- Agent/orchestrator retry semantics do not change; a permission-blocked tool call remains an ordinary failed tool observation
