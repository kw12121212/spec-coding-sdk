# Tasks: builtin-tool-manager

## Implementation

- [x] Create the built-in external tool manager package and define the public
      resolution/install contract for supported tools
- [x] Implement supported-tool detection for `rg`, `gh`, and `rtk`, preferring
      host-installed executables when present
- [x] Implement managed prebuilt-binary download and installation into an
      SDK-controlled directory
- [x] Implement reuse and failure handling so partial installs are not returned
      as valid executables

## Testing

- [x] Unit tests: resolve a host-installed supported tool without downloading
- [x] Unit tests: install a missing supported tool from a prebuilt artifact
- [x] Unit tests: reuse an existing managed installation without reinstalling
- [x] Unit tests: unsupported tool name returns an error
- [x] Unit tests: unsupported platform returns an error
- [x] Unit tests: failed download or failed install does not return a usable
      executable path
- [x] Lint passes
- [x] Unit tests pass

## Verification

- [x] Verify implementation matches proposal scope and leaves `tool-gh` /
      `tool-rtk` for later planned changes
