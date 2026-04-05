# Design: core-interfaces

## Approach

1. Define all interfaces in `internal/core/` as a single file `interfaces.go` — keeps foundational types in one place for easy discovery.
2. Each interface gets a single Go file with its definition and any closely related types (e.g., `Tool` alongside `ToolResult`).
3. Add compile-time checks where appropriate (e.g., verify concrete types satisfy interfaces).
4. Write tests that demonstrate how other packages would implement these interfaces.

## Key Decisions

- **Interfaces in `internal/core/`, not `pkg/sdk/`**: M10's explicit goal is to build the public SDK facade on top of M1 interfaces. Placing interfaces in `pkg/` now would preempt M10's scope. `internal/core/` is the right home for foundational contracts that internal packages implement.
- **Single-method `Tool` interface (`Execute`)**: claw-code's tool surface shows that not every tool needs separate validation — bash executes directly, glob fails naturally on bad patterns. A single `Execute` method is sufficient. If validation becomes necessary later, it can be added via interface composition in the specific milestone that needs it (YAGNI).
- **`Agent` as a minimal interface**: Agent contract at this stage should cover session lifecycle (start, stop) and tool invocation. State machine details belong to M3.
- **`Event` as a struct, not interface**: Events carry data (type, payload, timestamp) — a concrete struct is more appropriate than an interface. The `Event` "interface" in the milestone description maps to a typed struct with JSON serialization support.
- **`PermissionProvider` as a single-method interface**: One `Check(operation, resource) error` method. Sufficient for M2 tool hooks and M5's full permission model.

## Alternatives Considered

- **Separate `Validate`/`Execute` on Tool**: adds complexity not justified by current tool designs. Rejected per YAGNI — can be added later via interface embedding.
- **Interfaces in `pkg/sdk/`**: would conflict with M10's planned scope. Rejected.
- **Interface per file vs. single file**: single file `interfaces.go` is more discoverable for 4-5 interfaces. Can be split later if it grows unwieldy.
- **`Event` as interface with multiple implementations**: over-engineering at this stage. A concrete struct with a type discriminator is simpler and sufficient for M1's needs. M3/M6/M7 can introduce specialized event types.
