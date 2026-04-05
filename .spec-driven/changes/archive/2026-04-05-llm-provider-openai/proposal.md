# llm-provider-openai

## What

实现 OpenAI Chat Completions API 兼容的 LLM provider，使 SDK 可通过 OpenAI 格式调用大语言模型。由于 OpenAI API 是行业事实标准，此实现同时覆盖 OpenAI 官方及所有 OpenAI 兼容供应商（DeepSeek、Moonshot、GLM 等，通过配置 `BaseURL`）。

## Why

M5 当前仅有 provider 接口定义（`llm-provider-interface` 已完成），尚无具体实现。OpenAI 格式覆盖最广的供应商群体，是第一个 provider 实现的自然选择。有了具体 provider 后，agent 编排循环（M4）才能真正端到端工作，后续的 streaming 和 token 计数也有真实调用基础。

## Scope

- 在 `internal/llm/` 中新增 `openai` 子包，实现 `Provider` 接口
- 同步调用（`Complete`）：发送 OpenAI Chat Completions 请求，解析 JSON 响应
- 流式调用（`Stream`）：解析 SSE 流，逐 chunk 回调
- 请求格式转换：内部 `Request` → OpenAI JSON 格式
- 响应格式转换：OpenAI JSON 响应 → 内部 `Response` / `StreamChunk`
- 工具调用（tool_calls）的双向格式适配
- 错误处理：HTTP 错误、API 错误响应、网络超时
- 可配置 `BaseURL`（支持 OpenAI 兼容供应商）
- 使用 mock HTTP server 的单元测试

## Unchanged Behavior

- `internal/llm/` 包已有的类型和接口不变（Role、Message、Request、Response、StreamChunk、StreamCallback、Provider、ProviderConfig）
- `internal/agent/` 包的 Thinker 接口和编排逻辑不变
- 所有已有测试继续通过
