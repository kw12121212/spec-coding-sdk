# llm-token-counter

## What

在 `internal/llm/` 包中实现 token 计数估算和上下文窗口管理，使 agent 在发送 LLM 请求前能估算消息占用的 token 数，并判断是否超出模型的上下文窗口限制。

## Why

M5 的 Done Criteria 要求 "Token 计数可在发送前估算消息占用"。当前 provider 实现只能从 API 响应中获取实际 token 用量（`Usage`），无法在发送前预估。Agent 编排层需要在构建请求时判断消息是否超出窗口，以决定是否截断或压缩上下文。

## Scope

- 在 `internal/llm/` 包中定义 `EstimateTokens(messages []Message) int` 函数，基于字符启发式估算 token 数
- 在 `internal/llm/` 包中定义模型上下文窗口大小注册表（`ModelContextWindow`），覆盖主流 OpenAI 和 Claude 模型
- 在 `internal/llm/` 包中定义 `ContextChecker`，提供检查消息列表是否超出指定模型上下文窗口的方法
- 单元测试覆盖估算、注册表查询和上下文检查

## Unchanged Behavior

- 现有 `Provider` 接口及其所有实现（OpenAI、Claude）不变
- 现有 `Request`、`Response`、`Message`、`Usage` 等类型不变
- 现有流式处理（`streaming` 包）不变
