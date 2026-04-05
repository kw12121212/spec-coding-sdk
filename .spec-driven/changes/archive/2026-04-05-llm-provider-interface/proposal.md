# llm-provider-interface

## What

定义 provider 无关的 LLM 客户端抽象接口和统一消息类型，为后续 OpenAI、Claude 等具体 provider 实现提供契约基础。包含 `Provider` 接口、`LLMMessage`/`LLMRequest`/`LLMResponse` 等核心类型，以及流式回调类型定义。

## Why

M4（Agent 生命周期）已实现 `Thinker` 接口和编排循环，但目前只能用 mock 实现。需要定义 LLM 调用层的抽象接口，使 `Thinker` 的具体实现（M5 后续变更）能对接真实的 LLM API。统一的消息类型设计能让 OpenAI 和 Claude 两种 API 格式差异对上层透明。

## Scope

**In Scope:**
- `Provider` 接口定义（同步调用 + 流式调用）
- `LLMMessage` 统一消息类型（支持 user/assistant/tool 角色）
- `LLMRequest` 请求类型（模型名、消息列表、温度等参数）
- `LLMResponse` 响应类型（内容、工具调用、用量统计）
- `LLMToolCall` 工具调用类型（与 Orchestrator 的 `ToolCall` 对齐）
- `LLMUsage` 用量统计类型
- `StreamCallback` 流式回调函数类型
- `ProviderConfig` provider 配置基类
- 新增 `llm-backend.md` spec

**Out of Scope:**
- 具体 provider 实现（OpenAI、Claude）— 分别由 `llm-provider-openai` 和 `llm-provider-claude` 负责
- SSE 解析和流式响应处理 — 由 `llm-streaming` 负责
- Token 计数 — 由 `llm-token-counter` 负责
- `Thinker` 接口的修改 — 已在 M4 定义，本变更仅提供实现基础

## Unchanged Behavior

- 现有 `Thinker` 接口签名不变
- 现有 `ToolCall`、`ThinkResult` 类型不变
- Orchestrator 编排循环逻辑不变
- 所有已完成的工具和 agent 行为不变
