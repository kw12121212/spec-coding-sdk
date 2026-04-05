# agent-orchestrator

## What

实现 agent 的多轮工具调用编排循环（receive → think → act → observe），使 agent 能够接收用户消息、调用 LLM 获取决策、执行工具并将结果反馈给 LLM，直到 LLM 给出最终回复或达到最大迭代次数。

## Why

M4 的 agent-lifecycle 和 agent-conversation 已完成，agent 具备了状态管理和消息管理能力，但缺少核心编排逻辑——即驱动 think-act-observe 循环的引擎。没有编排器，agent 只能手动单步调用工具，无法自动完成多轮推理和工具调用。这是 M4 的最后一块拼图，完成后 M5（LLM 后端）和 M6（权限）可直接与之集成。

## Scope

**In Scope:**
- `Thinker` 接口定义（provider-agnostic LLM 调用抽象）
- `ToolCall` 类型定义（表示 LLM 返回的工具调用请求）
- `ThinkResult` 类型定义（表示 LLM 返回的完整响应）
- `Orchestrator` 结构体实现 think-act-observe 循环
- 循环最大迭代次数控制（默认值 50，可配置）
- 编排循环中的事件发射（每步发射对应事件）
- 使用 mock Thinker 和 stub Tool 的完整循环测试

**Out of Scope:**
- 真实 LLM provider 实现（M5）
- 权限检查注入（M6）
- 上下文压缩/截断策略（后续增强）
- 流式响应处理（M5）

## Unchanged Behavior

Behaviors that must not change as a result of this change (leave blank if nothing is at risk):
- BaseAgent 的现有生命周期方法（Start/Stop/Pause/Resume）行为不变
- Conversation 的消息管理行为不变
- core.Agent 接口定义不变
- 现有事件类型和结构不变
