# llm-backend — Delta

## ADDED Requirements

### Requirement: SSE 解析器

- 项目 MUST 在 `internal/llm/streaming/` 包中提供通用 SSE 流解析器 `SSEParser`。
- `SSEParser` MUST 从 `io.Reader` 读取 SSE 流并产出 `SSEEvent` 结构体（含 `Event` 和 `Data` 字段）。
- `SSEParser` MUST 支持 `event:` 行、`data:` 行和空行分隔的标准 SSE 格式。
- `SSEParser` MUST 在遇到 OpenAI 的 `data: [DONE]` 终止信号时返回 `io.EOF`。
- `SSEParser` MUST 在遇到 Claude 的 `event: message_stop` 信号时返回 `io.EOF`。
- `SSEParser` MUST 在遇到格式错误的行时返回错误。

### Requirement: 流式 Tool Call 累积器

- 项目 MUST 在 `internal/llm/streaming/` 包中提供 `ToolCallAccumulator`，自动拼接跨 chunk 的增量 tool call 数据。
- `ToolCallAccumulator.Feed(chunk StreamChunk)` MUST 按 tool call 的 index 缓存并拼接 partial JSON。
- 当某个 tool call 的 JSON 拼接完成时（累积器根据 provider 的终止信号判断），`Feed()` MUST 返回完整的 `ToolCall`。
- `ToolCallAccumulator.Flush()` MUST 返回所有当前累积的 tool call（含未完成的），供流结束时的错误处理或收尾使用。
- `ToolCallAccumulator` MUST 同时支持 OpenAI 格式（`tool_calls[i].function.arguments` 分片）和 Claude 格式（`input_json_delta` 分片）。

### Requirement: OpenAI Stream 使用公共 SSE 解析器

- `OpenAIProvider.Stream()` MUST 使用 `streaming.SSEParser` 解析 SSE 流，而非内嵌的 `bufio.Scanner` 行解析逻辑。
- `OpenAIProvider.Stream()` MUST 使用 `ToolCallAccumulator` 累积跨 delta chunk 的 tool call `function.arguments` 分片。
- 回调交付的 `StreamChunk` 中，`ToolCalls` 字段 MUST 包含已完整拼接的 `ToolCall`（有完整 `ID`、`Name`、`Input`）。

### Requirement: Claude Stream 使用公共 SSE 解析器

- `ClaudeProvider.Stream()` MUST 使用 `streaming.SSEParser` 解析 SSE 流，而非内嵌的 `bufio.Scanner` 行解析逻辑。
- `ClaudeProvider.Stream()` MUST 使用 `ToolCallAccumulator` 累积跨 `content_block_delta` 事件的 `input_json_delta` 分片。
- 回调交付的 `StreamChunk` 中，`ToolCalls` 字段 MUST 包含已完整拼接的 `ToolCall`。

### Requirement: SSE 解析器可测试性

- 测试 MUST 使用 `strings.Reader` 或 `bytes.Buffer` 构造 SSE 输入流。
- 测试 MUST 覆盖：标准 SSE 事件流、带 event 类型的流、`[DONE]` 终止、`message_stop` 终止、空行分隔、格式错误。

### Requirement: 累积器可测试性

- 测试 MUST 覆盖：单 chunk 完整 tool call、多 chunk 拼接完整 tool call、多个并发 tool call 的交错 chunk、flush 未完成数据。

## CHANGED Requirements

_无_
