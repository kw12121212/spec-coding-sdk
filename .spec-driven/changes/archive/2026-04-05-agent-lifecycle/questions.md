# Questions: agent-lifecycle

## Open

<!-- No open questions -->

## Resolved

- [x] Q: Should the agent state machine support custom intermediate states?
  Context: Determines whether to use extensible or fixed state definitions.
  A: Fixed state set (init/running/paused/stopped/error). No custom states — YAGNI.

- [x] Q: Should agent-lifecycle include event emission hooks (AgentStateEvent on transitions)?
  Context: Determines whether event emission is part of this change or deferred to agent-orchestrator.
  A: Include events now — emit AgentStateEvent on every state transition via optional EventEmitter.
