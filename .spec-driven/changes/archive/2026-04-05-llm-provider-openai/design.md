# Design: llm-provider-openai

## Approach

在 `internal/llm/openai/` 子包中实现 `Provider` 接口，使用 Go 标准库 `net/http` 发送 HTTP 请求。

### 核心结构

```
internal/llm/openai/
  provider.go      — OpenAIProvider struct, New() constructor, Provider 接口实现
  provider_test.go — 使用 httptest.Server 的单元测试
```

### 请求转换

内部 `llm.Request` 转换为 OpenAI Chat Completions API JSON 格式：
- `model` ← `Request.Model`（若空则使用 `ProviderConfig.Model`）
- `messages` — 按 Role 映射：
  - `RoleUser` → `{"role":"user","content":...}`
  - `RoleAssistant` → `{"role":"assistant","content":...,"tool_calls":[...]}`（如有）
  - `RoleTool` → `{"role":"tool","content":...,"tool_call_id":...}`
- `temperature` / `max_tokens` — 可选字段，零值时省略
- `stream` — `Complete` 时 false，`Stream` 时 true

### 响应转换（Complete）

OpenAI Chat Completions 响应中的 `choices[0]` 映射到内部 `Response`：
- `content` ← `message.content`
- `tool_calls` — 转换 OpenAI 格式（`function.name` + `function.arguments`）为内部 `ToolCall`
- `usage` → `Usage{PromptTokens, CompletionTokens}`
- `stop_reason` — `finish_reason` 映射：`"stop"` → `"stop"`, `"tool_calls"` → `"tool_use"`, `"length"` → `"max_tokens"`

### 响应转换（Stream）

解析 SSE 流（`data: ...` 行）：
- `choices[0].delta.content` → `StreamChunk.Content`
- `choices[0].delta.tool_calls` → `StreamChunk.ToolCalls`（增量格式适配）
- 最后一个 chunk 的 `usage` → `StreamChunk.Usage`
- `data: [DONE]` 结束流

### HTTP 请求

- `POST {BaseURL}/chat/completions`
- Headers: `Authorization: Bearer {APIKey}`, `Content-Type: application/json`
- 使用 `http.Client`，支持通过构造函数注入自定义 client（便于测试）

### 错误处理

- HTTP 非 2xx：解析 OpenAI 错误响应体，返回包含状态码和错误信息的 error
- 网络错误/超时：直接返回底层 error
- JSON 解析失败：包装原始错误返回

## Key Decisions

1. **使用标准库 `net/http`** — 不引入第三方 HTTP 客户端依赖（如 resty），保持依赖最小化
2. **子包而非同包文件** — `internal/llm/openai/` 独立子包，与未来 `internal/llm/claude/` 并列，避免单包膨胀
3. **构造函数注入 `http.Client`** — 通过可选参数注入，默认使用 `http.DefaultClient`，测试时注入 `httptest` client
4. **OpenAI 格式 tool_calls 适配** — OpenAI 使用嵌套 `function.name` / `function.arguments`，需扁平化为内部 `ToolCall`

## Alternatives Considered

- **使用官方 openai-go SDK** — 引入重量级依赖，且对兼容供应商的 base URL 配置不够透明，不适合作为内部 provider 实现
- **在同包内用文件区分** — `internal/llm/openai_provider.go` — 随 provider 增多文件会膨胀，子包更清晰
