# llm-backend delta for llm-provider-openai

## ADDED Requirements

### Requirement: OpenAI Provider 包结构

- 项目 MUST 在 `internal/llm/openai/` 包中提供 OpenAI API 兼容的 LLM provider 实现。
- `internal/llm/openai/` 包 MUST 实现 `internal/llm.Provider` 接口。

### Requirement: OpenAI Provider 构造

- 项目 MUST 提供 `NewProvider(config llm.ProviderConfig, opts ...Option) *OpenAIProvider` 构造函数。
- 项目 MUST 提供 `WithHTTPClient(client *http.Client) Option` 选项函数，允许注入自定义 HTTP client。
- 当未注入 `http.Client` 时，MUST 使用 `http.DefaultClient`。
- `OpenAIProvider` MUST 包含编译期接口满足检查 `var _ llm.Provider = (*OpenAIProvider)(nil)`。

### Requirement: OpenAI Provider Complete 方法

- `OpenAIProvider.Complete(ctx context.Context, req llm.Request)` MUST 向 `{BaseURL}/chat/completions` 发送 HTTP POST 请求，`stream` 字段设为 false。
- 请求 MUST 包含 `Authorization: Bearer {APIKey}` 和 `Content-Type: application/json` 头。
- 当 `req.Model` 为空时，MUST 使用 `ProviderConfig.Model`。
- 响应中的 `choices[0].message` MUST 转换为内部 `llm.Response`：
  - `content` → `Response.Content`
  - `tool_calls`（如有）→ 转换 OpenAI 格式（`function.name` + `function.arguments`）为 `llm.ToolCall`
  - `usage` → `llm.Usage`
  - `finish_reason` → `Response.StopReason`（映射：`"stop"` → `"stop"`, `"tool_calls"` → `"tool_use"`, `"length"` → `"max_tokens"`）
- HTTP 非 2xx 响应 MUST 返回错误（包含状态码和 OpenAI 错误信息）。

### Requirement: OpenAI Provider Stream 方法

- `OpenAIProvider.Stream(ctx context.Context, req llm.Request, callback llm.StreamCallback)` MUST 向 `{BaseURL}/chat/completions` 发送 HTTP POST 请求，`stream` 字段设为 true。
- MUST 逐行解析 SSE 流（`data: {json}` 行），将每个 chunk 转换为 `llm.StreamChunk` 并调用 `callback`。
- `data: [DONE]` 行 MUST 结束流。
- `callback` 返回非 nil error 时 MUST 停止流式推送并返回该 error。
- HTTP 非 2xx 响应 MUST 返回错误。

### Requirement: OpenAI 消息格式转换

- 内部 `llm.Message` 转换为 OpenAI 格式时 MUST 按角色映射：
  - `RoleUser` → `{"role":"user","content":...}`
  - `RoleAssistant` → `{"role":"assistant","content":...}` — 若有 `ToolCalls` 则附加 `tool_calls` 字段（转换为 OpenAI 嵌套 function 格式）
  - `RoleTool` → `{"role":"tool","content":...,"tool_call_id":...}`
- `Temperature` 和 `MaxTokens` 为零值时 MUST 从请求 JSON 中省略。

### Requirement: OpenAI Provider 可测试性

- 测试 MUST 使用 `net/http/httptest.Server` 模拟 OpenAI API。
- 测试 MUST 覆盖：同步调用正常响应、同步调用工具调用响应、流式调用、HTTP 错误响应、空消息列表。
