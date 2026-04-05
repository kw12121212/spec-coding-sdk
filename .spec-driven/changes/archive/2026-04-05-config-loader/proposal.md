# config-loader

## What

Implement YAML configuration file parsing for the SDK. Define a `Config` struct and a `LoadConfig` function that reads a YAML file and returns a validated configuration object.

## Why

M1 milestone requires configuration loading as a foundation type. Later milestones (LLM backend settings, permission policies, server listen addresses) will extend the config struct. Establishing the loading mechanism now ensures all subsequent features share a single config format and validation path.

## Scope

- Define `Config` struct in `internal/core/` with extensible top-level sections (initially empty/minimal)
- Implement `LoadConfig(path string) (*Config, error)` that reads and parses YAML
- Validate required fields and return descriptive errors on invalid input
- Unit tests covering happy path, missing file, malformed YAML, and validation errors

## Unchanged Behavior

- Existing `go build ./...` and `go test ./...` MUST continue to pass
- Existing interfaces (`Tool`, `Agent`, `Event`, `PermissionProvider`) MUST NOT be modified
