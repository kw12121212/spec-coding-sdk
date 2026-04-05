# spec-coding-sdk

Go SDK reimplementation of [claw-code](https://github.com/anthropics/claude-code) (a Rust coding agent harness), exposing its core capabilities through multiple interface layers.

## Architecture

The SDK provides four interface layers for third-party integration:

- **Native Go SDK** — direct library usage within Go applications
- **JSON-RPC** over stdin — subprocess integration
- **HTTP REST API** — service-oriented integration
- **gRPC** — high-performance RPC integration

## Features

- Tool surface: bash, file ops, grep, glob, LSP client, MCP protocol
- Agent lifecycle with structured event system
- Task, team, and cron registries
- Permission model with execution hooks
- LLM backend integration (Anthropic Claude, extensible)
- Built-in spec-driven development workflow

## Tech Stack

- Go 1.25+
- Protocol Buffers (gRPC)
- Apache 2.0 Licensed

## Development

This project uses a [spec-driven](/.spec-driven/) development workflow. See [`.spec-driven/roadmap/INDEX.md`](/.spec-driven/roadmap/INDEX.md) for the milestone plan.

### Prerequisites

- Go 1.25+
- protoc (for gRPC code generation)

### Build & Test

```bash
go build ./...
go test ./...
```

## Project Status

Early development — core interfaces and project scaffold in progress. See the [roadmap](/.spec-driven/roadmap/INDEX.md) for planned milestones.

## License

[Apache License 2.0](LICENSE)
