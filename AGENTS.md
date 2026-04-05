# AGENTS.md

Guidelines for AI coding agents working on this project.

## Project Overview

This is a Go SDK reimplementation of claw-code (a Rust coding agent harness). It exposes coding agent capabilities through four interface layers: native Go SDK, JSON-RPC, HTTP REST, and gRPC. The project uses a spec-driven development workflow managed under `.spec-driven/`.

## Essential Reading

Before making any changes, read these files:

1. `.spec-driven/specs/INDEX.md` — current spec index
2. `.spec-driven/config.yaml` — project context and rules
3. `.spec-driven/roadmap/INDEX.md` — milestone plan and progress

## Development Rules

- **Spec-first**: Only implement what is described in specs. If scope needs to expand, update specs first via `/spec-driven-modify`.
- **Observable behavior**: Specs describe observable behavior only. Tests verify behavior, not implementation details.
- **No speculative code**: Implement only what the current task requires (YAGNI). No abstractions for hypothetical future needs.
- **Read before modify**: Always read existing code before modifying it.
- **Test requirements**: Every change must include tests (lint + unit tests minimum). Each test must be independent — no shared mutable state. Prefer real dependencies over mocks for code the project owns.
- **MUST/SHOULD/MAY**: Respect requirement strength. MUST = required, SHOULD = default unless justified, MAY = optional.

## Code Style

- Follow standard Go conventions (`gofmt`, `golint`)
- Go module at repository root
- Protobuf-generated `*.pb.go` files are build artifacts

## Workflow

Use spec-driven skills for all changes:

- `/spec-driven-propose` — propose a new change
- `/spec-driven-apply` — implement tasks
- `/spec-driven-verify` — verify completion
- `/spec-driven-review` — code quality review
- `/spec-driven-archive` — archive completed change

## Language

Specs and roadmap are written in Chinese. Code, comments, and commit messages should be in English.
