# llm-provider-claude — Tasks

## Implementation

- [x] Create `internal/llm/claude/` package with `provider.go`
- [x] Define Claude wire types (request/response structs for Messages API)
- [x] Implement `NewProvider` constructor with `WithHTTPClient` option
- [x] Implement format conversion: `llm.Message` → Claude request format
  - System message extraction to top-level `system` field
  - User/assistant/tool role mapping to Claude content blocks
  - Tool call ↔ `tool_use`/`tool_result` content block conversion
- [x] Implement format conversion: Claude response → `llm.Response`
  - Content block array concatenation for text
  - `tool_use` block → `llm.ToolCall` mapping
  - `stop_reason` mapping (`end_turn`→`stop`, `tool_use`→`tool_use`, `max_tokens`→`max_tokens`)
- [x] Implement `Complete` method
  - HTTP POST to `{BaseURL}/v1/messages` with `x-api-key`, `anthropic-version` headers
  - `max_tokens` default to 4096 when zero
  - Error handling for non-2xx responses
- [x] Implement `Stream` method
  - HTTP POST with `stream: true`
  - SSE parsing for typed events (`content_block_delta`, `message_delta`, `message_stop`)
  - `text_delta` → `StreamChunk.Content`
  - `input_json_delta` → `StreamChunk.ToolCalls`
  - `message_delta` usage → final `StreamChunk.Usage`
  - Callback error propagation

## Testing

- [x] Test `Complete` with normal text response
- [x] Test `Complete` with tool_use response
- [x] Test `Stream` with text chunks
- [x] Test `Stream` with tool call chunks
- [x] Test HTTP error response (non-2xx)
- [x] Test empty messages list
- [x] Test model fallback to config default
- [x] Test message format conversion (all role types including system)
- [x] Test `max_tokens` default when zero
- [x] Test callback error stops streaming

## Verification

- [x] `go vet ./internal/llm/claude/...` passes
- [x] `go test ./internal/llm/claude/...` passes
- [x] All delta spec requirements are independently testable
