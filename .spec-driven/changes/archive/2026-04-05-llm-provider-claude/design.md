# llm-provider-claude — Design

## Approach

Follow the same structural pattern as `internal/llm/openai/`:

1. **Wire types** — define Claude-specific request/response structs as unexported types in `internal/llm/claude/`
2. **Format conversion** — convert between `llm.*` types and Claude wire types in helper functions
3. **HTTP transport** — share the same `newHTTPRequest` / `parseAPIError` pattern from the OpenAI provider
4. **SSE parsing** — Claude streaming uses typed events (`event: content_block_delta`, etc.), parse with a scanner that reads both `event:` and `data:` lines

### Key structural differences from OpenAI

| Aspect | OpenAI | Claude |
|--------|--------|--------|
| Endpoint | `/chat/completions` | `/v1/messages` |
| Auth header | `Authorization: Bearer {key}` | `x-api-key: {key}` |
| Version header | none | `anthropic-version: 2023-06-01` |
| System message | mixed into messages array | separate `system` field |
| Response content | flat `content` string | array of content blocks (`text`, `tool_use`) |
| Tool call location | `tool_calls` on message | `tool_use` content block in `content[]` |
| Tool result | `role: "tool"` message | `role: "user"` with `tool_result` content block |
| stop_reason values | `stop`, `tool_calls`, `length` | `end_turn`, `tool_use`, `max_tokens` |
| max_tokens | optional | required |
| SSE events | `data: {json}` only | typed events: `message_start`, `content_block_start`, `content_block_delta`, etc. |

## Key Decisions

1. **System message handling** — Messages with role `"system"` (as a raw string) will be extracted into Claude's top-level `system` field and omitted from the messages array. This avoids modifying the shared `llm` package to add `RoleSystem`.

2. **max_tokens default** — Claude requires `max_tokens`. When `llm.Request.MaxTokens` is 0, default to 4096.

3. **Content block handling** — Claude's response `content` is an array of blocks. Each block is either `{"type":"text","text":"..."}` or `{"type":"tool_use","id":"...","name":"...","input":{...}}`. Concatenate all `text` blocks into `Response.Content` and collect all `tool_use` blocks into `Response.ToolCalls`.

4. **Tool result mapping** — Claude sends tool results as `role: "user"` messages with `tool_result` content blocks. When converting from `llm.Message{Role: RoleTool}`, emit a `role: "user"` message with `tool_result` content block.

5. **Streaming model** — Claude SSE has typed events. Parse `event:` line then `data:` line. Key events:
   - `message_start`: contains usage (input tokens)
   - `content_block_start`: new content block (text or tool_use)
   - `content_block_delta`: incremental content (`text_delta` or `input_json_delta`)
   - `message_delta`: contains `stop_reason` and output usage
   - `message_stop`: end of stream

## Alternatives Considered

- **Common SSE parser shared with OpenAI** — rejected; Claude's typed event model is fundamentally different from OpenAI's flat `data:` lines. Sharing a parser would add complexity without benefit.
- **Modifying `llm` package to add `RoleSystem`** — rejected; keeps this change self-contained. System prompt can be handled by convention during format conversion.
