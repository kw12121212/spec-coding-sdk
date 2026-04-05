# M5 - LLM 后端集成

## Goal

实现 LLM 调用层，为 agent 的 think 步骤提供大语言模型推理能力，支持多 provider 和流式响应。

## In Scope

- LLM 客户端抽象接口（provider-agnostic）
- Anthropic Claude API 适配（首要 provider）
- 流式响应处理（SSE / streaming）
- Token 计数与上下文窗口管理
- 请求重试与错误处理

## Out of Scope

- 非 Anthropic provider 的实现（可后续扩展）
- 上下文压缩/截断策略（属于会话管理增强）
- Agent 编排逻辑（M4）

## Done Criteria

- LLM 客户端可成功调用 Anthropic API 并获得完整响应
- 流式响应可逐 token 回调到调用方
- Token 计数可在发送前估算消息占用
- 请求失败时有合理的重试和错误上报
- 有单元测试覆盖（使用 mock HTTP server）

## Planned Changes

- `llm-client` - Declared: planned - LLM 客户端抽象接口与 Anthropic provider 实现
- `llm-streaming` - Declared: planned - 流式响应处理与回调机制实现
- `llm-token-counter` - Declared: planned - Token 计数与上下文窗口管理实现

## Dependencies

- M1 核心接口（Event 接口用于流式事件）

## Risks

- Anthropic API 版本升级可能导致适配层变更
- Token 计数的精确性受 tokenizer 实现影响
- 流式响应的错误恢复和中断处理复杂度

## Status

- Declared: proposed

## Notes

- LLM 客户端接口设计应支持未来添加其他 provider（OpenAI、本地模型等）
- 与 M2、M4 可并行开发；M4 使用 mock LLM 完成编排测试，M5 提供真实实现后两者集成
- 参考 claw-code 的 LLM 调用模式，但抽象层设计由本 SDK 自行决定
