# llm-backend delta: llm-token-counter

## ADDED Requirements

### Requirement: Token 估算函数

- 项目 MUST 在 `internal/llm/` 包中提供 `EstimateTokens(messages []Message) int` 函数。
- `EstimateTokens` MUST 基于字符启发式估算消息列表的 token 数（约 4 个字符 = 1 token）。
- `EstimateTokens` MUST 将每条 `Message` 的 `Content` 字段和 `ToolCalls` 的 JSON 序列化结果都纳入估算。
- `EstimateTokens` 对空消息列表 MUST 返回 0。

### Requirement: 模型上下文窗口注册表

- 项目 MUST 在 `internal/llm/` 包中定义 `ModelContextWindow` 结构体，包含以下字段：
  - `ModelID` (string) — 模型标识
  - `TotalTokens` (int) — 上下文窗口总 token 数
- 项目 MUST 在 `internal/llm/` 包中提供 `ContextWindow(modelID string) (int, bool)` 函数。
- `ContextWindow` 对已知模型 MUST 返回对应的 `TotalTokens` 和 `true`。
- `ContextWindow` 对未知模型 MUST 返回 `0, false`。
- 注册表 MUST 至少覆盖以下模型的上下文窗口大小：gpt-4o、gpt-4-turbo、gpt-3.5-turbo、claude-sonnet-4-6、claude-opus-4-6、claude-haiku-4-5。

### Requirement: 上下文窗口检查

- 项目 MUST 在 `internal/llm/` 包中定义 `ContextChecker` 结构体。
- `ContextChecker` MUST 提供 `Fits(messages []Message, model string, reserved int) bool` 方法：
  - 使用 `EstimateTokens` 估算消息 token 数
  - 使用 `ContextWindow` 获取模型窗口大小
  - 返回 `估算值 + reserved <= 窗口大小`
  - 当模型未知时 MUST 返回 `false`
- `ContextChecker` MUST 提供 `Remaining(messages []Message, model string, reserved int) (int, error)` 方法：
  - 返回 `窗口大小 - 估算值 - reserved`
  - 当模型未知时 MUST 返回错误
- `reserved` 参数表示为响应预留的 token 数（如 `max_tokens`）。

### Requirement: Token 计数可测试性

- 测试 MUST 覆盖：`EstimateTokens` 对空列表、纯文本、含 tool call、多消息混合的估算。
- 测试 MUST 覆盖：`ContextWindow` 对已知模型和未知模型的查询。
- 测试 MUST 覆盖：`ContextChecker` 的 fits/remaining 在正常、超窗口、reserved、未知模型场景。
