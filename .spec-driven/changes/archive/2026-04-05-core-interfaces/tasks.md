# Tasks: core-interfaces

## Implementation

- [x] Define `Tool` interface with `Execute` method and `ToolResult` return type in `internal/core/`
- [x] Define `Agent` interface with session lifecycle methods (`Start`, `Stop`) in `internal/core/`
- [x] Define `Event` struct with type, payload, and timestamp fields in `internal/core/`
- [x] Define `PermissionProvider` interface with `Check` method in `internal/core/`
- [x] Verify `go build ./...` passes with all new types

## Testing

- [x] Write tests verifying `Tool` interface can be implemented by a concrete type in a separate test package
- [x] Write tests verifying `Agent` interface can be implemented by a concrete type
- [x] Write tests verifying `Event` struct can be instantiated and its fields accessed
- [x] Write tests verifying `PermissionProvider` interface can be implemented
- [x] `go test ./...` passes
- [x] `make lint` passes

## Verification

- [x] All core interfaces (`Tool`, `Agent`, `PermissionProvider`) are defined and referenceable
- [x] `Event` struct is defined with required fields
- [x] `ToolResult` type is defined
- [x] No changes to `pkg/sdk/` or `api/proto/`
- [x] Existing `go build ./...`, `go test ./...`, `make lint` still pass
