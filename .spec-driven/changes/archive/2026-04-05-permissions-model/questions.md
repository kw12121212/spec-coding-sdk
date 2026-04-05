# Questions: permissions-model

## Open

<!-- No open questions -->

## Resolved

- [x] Q: Should DefaultPolicy's "deny destructive bash commands" use a hardcoded pattern list or a configurable one?
  Context: A hardcoded list (rm -rf, mkfs, dd, etc.) is simpler but inflexible. A configurable list can be tuned per deployment but adds API surface.
  A: Hardcoded pattern list.

- [x] Q: Should the DefaultPolicy's "write outside CWD" restriction use the process working directory or an explicitly set root?
  Context: Process CWD is implicit and may change. An explicit root is clearer but requires configuration.
  A: Use CWD at policy creation time.

- [x] Q: What should DefaultPolicy's default access level be?
  Context: Restrictive vs permissive default affects SDK ergonomics and safety.
  A: Full access — DefaultPolicy() allows everything by default (DefaultDecision=DecisionAllow, no restrictive rules).
