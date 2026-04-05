# agent-lifecycle

## What

实现 agent 状态机与生命周期管理。在 `internal/agent/` 包中提供 `BaseAgent` 结构体，实现 `internal/core.Agent` 接口，支持固定状态集合（init → running → paused → stopped → error）之间的合法转换，并在每次状态转换时通过 `EventEmitter` 发出 `AgentStateEvent`。

## Why

M1 定义了 `Agent` 接口（Start/Stop/RunTool），但没有提供实现。M4 的 agent 生命周期是整个系统的核心运行时基础——后续的会话管理（agent-conversation）和多轮编排循环（agent-orchestrator）都依赖于 agent 拥有正确的状态管理和生命周期控制。尽早完成此变更可解除 M5（LLM 后端）、M6（权限）、M7（注册表）等 4+ 下游里程碑的依赖。

## Scope

**In scope:**
- 定义 agent 状态类型和合法转换规则（init → running, running → paused, paused → running, running → stopped, * → error）
- 实现 `BaseAgent` 结构体，满足 `core.Agent` 接口
- `Start()` 和 `Stop()` 方法触发状态转换
- 每次状态转换通过可选的 `EventEmitter` 发出 `AgentStateEvent`
- `RunTool()` 在 running 状态下执行工具调用；非 running 状态返回错误
- 状态转换的并发安全

**Out of scope:**
- 会话与消息管理（agent-conversation）
- 多轮编排循环（agent-orchestrator）
- LLM 调用集成（M5）
- 权限策略实现（M6）
- 自定义/可扩展状态

## Unchanged Behavior

- `core.Agent` 接口签名不变（Start/Stop/RunTool）
- 已有的 core 类型（ToolResult、Event、EventEmitter 等）不变
- 已实现的工具（bash、fileops、grep、glob）行为不变
