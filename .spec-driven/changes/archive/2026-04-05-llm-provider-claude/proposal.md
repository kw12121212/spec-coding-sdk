# llm-provider-claude

## What

Implement an Anthropic Claude Messages API provider in `internal/llm/claude/` that satisfies the `llm.Provider` interface, supporting synchronous `Complete`, streaming `Stream`, and Claude-specific message format conversion.

## Why

M5 (LLM 后端集成) already has the provider-agnostic interface (`llm.Provider`) and an OpenAI-compatible provider. The Claude Messages API is the second required provider. Both providers must exist before the streaming unification (`llm-streaming`) and token counting (`llm-token-counter`) changes can provide cross-provider value. The Claude API is structurally different from OpenAI (content blocks, separate system field, different SSE event types), so a dedicated provider is necessary.

## Scope

- `internal/llm/claude/` package implementing `llm.Provider`
- Claude Messages API wire types (`/v1/messages` endpoint)
- Message format conversion: `llm.Message` ↔ Claude content blocks
- `Complete` method: synchronous call with text and tool_use response handling
- `Stream` method: SSE parsing for Claude streaming events (`message_start`, `content_block_start`, `content_block_delta`, `content_block_stop`, `message_delta`, `message_stop`)
- Tool call handling: Claude `tool_use`/`tool_result` content blocks ↔ `llm.ToolCall`
- `x-api-key` and `anthropic-version` header handling
- `max_tokens` as required field (Claude API requires it)
- Unit tests using `net/http/httptest.Server`
- Delta spec additions to `llm-backend.md`

## Unchanged Behavior

- `internal/llm/` package types and `Provider` interface remain unchanged
- `internal/llm/openai/` package remains unchanged
- No changes to agent, tool surface, or any other package
