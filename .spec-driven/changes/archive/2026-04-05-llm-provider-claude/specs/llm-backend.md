# llm-backend

## ADDED Requirements

### Requirement: Claude Provider 包结构

- 项目 MUST 在 `internal/llm/claude/` 包中提供 Anthropic Claude Messages API 的 LLM provider 实现。
- `internal/llm/claude/` 包 MUST 实现 `internal/llm.Provider` 接口。

### Requirement: Claude Provider 构造

- 项目 MUST 提供 `NewProvider(config llm.ProviderConfig, opts ...Option) *ClaudeProvider` 构造函数。
- 项目 MUST 提供 `WithHTTPClient(client *http.Client) Option` 选项函数，允许注入自定义 HTTP client。
- 当未注入 `http.Client` 时，MUST 使用 `http.DefaultClient`。
- `ClaudeProvider` MUST 包含编译期接口满足检查 `var _ llm.Provider = (*ClaudeProvider)(nil)`。

### Requirement: Claude Provider Complete 方法

- `ClaudeProvider.Complete(ctx context.Context, req llm.Request)` MUST 向 `{BaseURL}/v1/messages` 发送 HTTP POST 请求。
- 请求 MUST 包含 `x-api-key: {APIKey}`、`Content-Type: application/json`、`anthropic-version: 2023-06-01` 头。
- 当 `req.Model` 为空时，MUST 使用 `ProviderConfig.Model`。
- 当 `req.MaxTokens` 为 0 时，MUST 使用默认值 4096 作为 `max_tokens`。
- 响应中 `content` 数组中的 `text` 类型块 MUST 拼接为 `Response.Content`。
- 响应中 `content` 数组中的 `tool_use` 类型块 MUST 转换为 `llm.ToolCall`（`name` → `Name`，`id` → `ID`，`input` → `Input`）。
- `stop_reason` MUST 映射：`"end_turn"` → `"stop"`，`"tool_use"` → `"tool_use"`，`"max_tokens"` → `"max_tokens"`。
- HTTP 非 2xx 响应 MUST 返回错误（包含状态码和 Claude 错误信息）。

### Requirement: Claude Provider Stream 方法

- `ClaudeProvider.Stream(ctx context.Context, req llm.Request, callback llm.StreamCallback)` MUST 向 `{BaseURL}/v1/messages` 发送 HTTP POST 请求，`stream` 字段设为 true。
- MUST 逐行解析 SSE 流，处理以下事件类型：
  - `content_block_delta` 中的 `text_delta` → `StreamChunk.Content`
  - `content_block_delta` 中的 `input_json_delta` → 追加到 `StreamChunk.ToolCalls`
  - `message_delta` 中的 `stop_reason` 和 `usage` → 最终 chunk 的 `Usage`
- `event: message_stop` MUST 结束流。
- `callback` 返回非 nil error 时 MUST 停止流式推送并返回该 error。
- HTTP 非 2xx 响应 MUST 返回错误。

### Requirement: Claude 消息格式转换

- 内部 `llm.Message` 转换为 Claude 格式时 MUST 按角色映射：
  - `RoleUser`（无 ToolCallID）→ `{"role":"user","content":[{"type":"text","text":...}]}`
  - `RoleUser`（有 ToolCallID）→ `{"role":"user","content":[{"type":"tool_result","tool_use_id":...,"content":...}]}`
  - `RoleAssistant`（有 ToolCalls）→ `{"role":"assistant","content":[{"type":"text","text":...}, {"type":"tool_use","id":...,"name":...,"input":...}]}`
  - `RoleAssistant`（无 ToolCalls）→ `{"role":"assistant","content":[{"type":"text","text":...}]}`
  - role 为 `"system"` 的消息 → 提取到 Claude 请求的 `system` 顶层字段
- `Temperature` 为零值时 MUST 从请求 JSON 中省略。

### Requirement: Claude Provider 可测试性

- 测试 MUST 使用 `net/http/httptest.Server` 模拟 Claude Messages API。
- 测试 MUST 覆盖：同步调用文本响应、同步调用工具调用响应、流式调用文本响应、流式调用工具调用响应、HTTP 错误响应、空消息列表。
