# Design: config-loader

## Approach

Add a `config.go` file in `internal/core/` that defines the `Config` struct and the `LoadConfig` function. Use the standard `gopkg.in/yaml.v3` library for YAML parsing. Keep the initial `Config` struct minimal — only fields that are needed now — so later milestones extend it without migration.

### Config struct layout

```go
type Config struct {
    // Fields will be added by later milestones (LLM, permissions, server, etc.)
}
```

The struct starts empty. Each future milestone adds its own section. This avoids premature field design while establishing the loading and validation pattern early.

### Loading flow

1. Read file contents from the given path
2. Parse YAML into `Config` struct
3. Run validation (initially: file existence and valid YAML syntax)
4. Return validated `*Config` or a descriptive error

## Key Decisions

- **`gopkg.in/yaml.v3` over `encoding/json`**: The project specification names config.yaml, so YAML is the canonical format. v3 is the latest stable API.
- **Empty initial struct**: Rather than guess future fields, start minimal. Each milestone adds its own section with a delta spec update.
- **Validation at load time**: Catch configuration errors early, before any component starts. Return structured errors with the field path and reason.

## Alternatives Considered

- **Viper**: Rejected — adds a large dependency for functionality that a 30-line loader provides. Can reconsider if config needs grow to include env var merging, remote config, etc.
- **TOML/JSON config**: The milestone scope specifies config.yaml; YAML is the target format.
- **Config in `pkg/`**: The config struct is internal plumbing; external consumers get it through the SDK facade (M10), not directly.
