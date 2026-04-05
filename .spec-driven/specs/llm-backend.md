# llm-backend

## ADDED Requirements

### Requirement: LLM 包结构

- 项目 MUST 在 `internal/llm/` 包中提供 LLM 调用层的所有抽象类型和接口。
- `internal/llm/` 包 MUST 可被 `internal/agent/` 及其他内部包引用。

### Requirement: LLM 角色类型

- 项目 MUST 在 `internal/llm/` 包中定义 `Role` 类型（基于 string），包含以下角色常量：`RoleUser`（"user"）、`RoleAssistant`（"assistant"）、`RoleTool`（"tool"）。

### Requirement: LLM 工具调用类型

- 项目 MUST 在 `internal/llm/` 包中定义 `ToolCall` 结构体，包含以下字段：
  - `ID` (string) — 工具调用的唯一标识（由 LLM 返回）
  - `Name` (string) — 工具名称
  - `Input` (json.RawMessage) — 工具输入参数
- `ToolCall` MUST 可 JSON 序列化/反序列化。

### Requirement: LLM 消息类型

- 项目 MUST 在 `internal/llm/` 包中定义 `Message` 结构体，包含以下字段：
  - `Role` (Role) — 消息角色
  - `Content` (string) — 消息内容
  - `ToolName` (string) — 工具名称（仅 tool 角色使用，其他角色为零值）
  - `ToolCallID` (string) — 工具调用 ID（仅 tool 角色使用，与 ToolCall.ID 对应）
  - `ToolCalls` ([]ToolCall) — 工具调用列表（仅 assistant 角色使用，其他角色为零值）
- `Message` MUST 可 JSON 序列化/反序列化（round-trip 一致）。

### Requirement: LLM 请求类型

- 项目 MUST 在 `internal/llm/` 包中定义 `Request` 结构体，包含以下字段：
  - `Model` (string) — 模型标识
  - `Messages` ([]Message) — 消息列表
  - `Temperature` (float64) — 采样温度（可选，0 表示使用模型默认值）
  - `MaxTokens` (int) — 最大生成 token 数（可选，0 表示使用模型默认值）
- `Request` MUST 可 JSON 序列化/反序列化。

### Requirement: LLM 响应类型

- 项目 MUST 在 `internal/llm/` 包中定义 `Response` 结构体，包含以下字段：
  - `Content` (string) — LLM 生成的文本内容
  - `ToolCalls` ([]ToolCall) — LLM 请求的工具调用列表（可为空）
  - `Usage` (Usage) — token 用量统计
  - `StopReason` (string) — 停止原因（如 "stop"、"tool_use"、"max_tokens"）
- `Response` MUST 可 JSON 序列化/反序列化。

### Requirement: LLM 用量统计类型

- 项目 MUST 在 `internal/llm/` 包中定义 `Usage` 结构体，包含以下字段：
  - `PromptTokens` (int) — 输入 token 数
  - `CompletionTokens` (int) — 输出 token 数
- `Usage` MUST 可 JSON 序列化/反序列化。

### Requirement: 流式回调类型

- 项目 MUST 在 `internal/llm/` 包中定义 `StreamChunk` 结构体，包含以下字段：
  - `Content` (string) — 增量文本内容
  - `ToolCalls` ([]ToolCall) — 增量工具调用（流式场景下 MAY 为部分数据）
  - `Usage` (Usage) — 用量统计（MAY 仅在最后一个 chunk 中有值）
- 项目 MUST 在 `internal/llm/` 包中定义 `StreamCallback` 函数类型为 `func(chunk StreamChunk) error`。
- 当 `StreamCallback` 返回非 nil error 时，provider 实现 MUST 停止流式推送。

### Requirement: Provider 接口

- 项目 MUST 在 `internal/llm/` 包中定义 `Provider` 接口，包含以下方法：
  - `Complete(ctx context.Context, req Request) (Response, error)` — 同步调用，返回完整响应
  - `Stream(ctx context.Context, req Request, callback StreamCallback) error` — 流式调用，逐 chunk 回调
- `Provider` 接口 MUST 可由外部包实现（不依赖 `internal/llm/` 的非导出类型）。

### Requirement: Provider 配置类型

- 项目 MUST 在 `internal/llm/` 包中定义 `ProviderConfig` 结构体，包含以下字段：
  - `BaseURL` (string) — API 基础 URL
  - `APIKey` (string) — API 密钥
  - `Model` (string) — 默认模型标识
- `ProviderConfig` MUST 可 JSON 序列化/反序列化。

### Requirement: 接口可实现性验证

- 测试 MUST 验证外部包可以定义实现 `Provider` 接口的具体类型。

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
