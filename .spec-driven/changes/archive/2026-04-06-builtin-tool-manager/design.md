# Design: builtin-tool-manager

## Approach

Add a dedicated external tool management package that resolves supported tool
names into executable paths. Resolution first checks the host environment for an
existing binary and returns it immediately when present. If the tool is missing,
the manager downloads a supported prebuilt archive, installs it into an
SDK-managed directory, and then returns the installed executable path together
with source metadata.

The change is intentionally limited to the manager contract and its observable
install behavior. Higher-level wrappers such as `tool-gh` and `tool-rtk` remain
separate roadmap items, and the proposal does not change the existing tool
surface contracts.

## Key Decisions

1. Prefer system-installed binaries before managed downloads.
   Rationale: this minimizes network dependency, respects existing user
   environments, and keeps the managed installer as a fallback rather than the
   default path.

2. Restrict installation to prebuilt binaries only.
   Rationale: the roadmap explicitly excludes source builds, and prebuilt assets
   keep installation behavior predictable across supported platforms.

3. Install into an SDK-managed directory instead of mutating global PATH
   locations.
   Rationale: this keeps the change self-contained, avoids privileged writes,
   and makes cleanup and reuse easier to reason about.

4. Keep the manager separate from tool wrapper behavior.
   Rationale: `builtin-tool-manager`, `tool-gh`, and `tool-rtk` are separate
   planned changes. Combining them here would expand scope beyond the selected
   roadmap item.

5. Treat failed installs as non-reusable.
   Rationale: callers need a clear guarantee that any returned path points to a
   complete executable, not a partially written artifact.

## Alternatives Considered

- Build missing tools from source during installation.
  Rejected because the roadmap explicitly limits external tool integration to
  prebuilt binaries.

- Vendor all supported binaries in the repository.
  Rejected because it would bloat the repository and complicate platform
  coverage for artifacts that change independently of the SDK.

- Fold `tool-gh` and `tool-rtk` into this same proposal.
  Rejected because those are already separate planned changes and would turn a
  foundation change into a multi-feature milestone bundle.
