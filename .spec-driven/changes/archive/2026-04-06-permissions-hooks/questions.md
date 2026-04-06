# Questions: permissions-hooks

## Open

<!-- No open questions -->

## Resolved

- [x] Q: Should `permissions-hooks` include human confirmation UX or only define blocked behavior for non-interactive execution paths?
  Context: `DecisionNeedConfirmation` exists in the permission model, but the current execution path exposes only `core.PermissionProvider.Check(... ) error`. Pulling UX into this change would expand M6 into SDK and transport design.
  A: Only define internal interception and blocked-execution behavior. Interface-layer confirmation UX stays out of scope for this change.
