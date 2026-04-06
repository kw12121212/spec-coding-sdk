# builtin-tool-manager

## What

Define the M3 built-in external tool manager that detects existing installations of
supported external tools and installs managed prebuilt binaries when they are
missing. The initial scope covers the manager behavior for `rg`, `gh`, and `rtk`
without implementing the higher-level `tool-gh` and `tool-rtk` wrappers.

## Why

The roadmap's next unfinished dependency starts at M3. The current grep tool spec
assumes `rg` is available, but the roadmap explicitly assigns prebuilt binary
management to the built-in tool manager. Defining this behavior now closes that
gap and creates the foundation required by later M3 changes.

## Scope

In scope:
- Detect whether `rg`, `gh`, or `rtk` is already available on the host system
- Resolve a usable executable path for a requested supported tool
- Download and install a supported prebuilt binary into an SDK-managed location
  when the tool is missing
- Reuse an existing managed installation instead of reinstalling the same tool
- Return observable metadata about whether the resolved executable came from the
  system or from the managed installation
- Add unit-test coverage for detection, install, reuse, unsupported tools, and
  failure cases

Out of scope:
- Implementing the `tool-gh` command wrapper
- Implementing the `tool-rtk` command wrapper
- Changing agent lifecycle, permissions, or interface-layer behavior
- Building tools from source instead of using prebuilt binaries

## Unchanged Behavior

Behaviors that must not change as a result of this change (leave blank if nothing is at risk):
- Existing `tool-bash`, `tool-file-ops`, `tool-glob`, and `tool-grep` contracts
  remain unchanged in this proposal
- Existing permission model and agent orchestration behavior remain unchanged
