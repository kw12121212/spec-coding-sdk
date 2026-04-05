# Design: agent-lifecycle

## Approach

在 `internal/agent/` 包中创建 `BaseAgent` 结构体，实现 `core.Agent` 接口。使用固定状态枚举和显式的转换表来管理状态机。通过 `sync.RWMutex` 保护状态字段的并发访问。

**核心结构：**

```
internal/agent/
  agent.go      — BaseAgent 结构体、状态类型定义、转换逻辑
  agent_test.go — 单元测试
```

**状态机模型：**

状态枚举：`StateInit`, `StateRunning`, `StatePaused`, `StateStopped`, `StateError`

合法转换表：
- init → running（Start）
- running → paused（Pause，新增方法但不属于 core.Agent 接口）
- paused → running（Resume，新增方法）
- running → stopped（Stop）
- paused → stopped（Stop）
- * → error（内部错误触发）
- error → stopped（Stop 可从 error 状态调用）

**构造函数：** `New(opts ...Option) *BaseAgent`，通过 functional options 注入 EventEmitter。

**RunTool 实现：** 仅在 StateRunning 状态下委托给 Tool.Execute，否则返回错误结果。

## Key Decisions

1. **固定状态集合** — 不支持自定义状态。原因：YAGNI，当前所有下游里程碑（M5-M14）的需求均可由这五个状态覆盖。
2. **EventEmitter 在构造时注入** — 而非每次调用传入。原因：agent 的生命周期事件是内部行为，不应暴露为调用方参数。
3. **Pause/Resume 不在 core.Agent 接口中** — 它们是 BaseAgent 的扩展方法。原因：core.Agent 接口在 M1 中已定义为 Start/Stop/RunTool，不修改已有接口。调用方如需 Pause/Resume 可通过类型断言或 future 接口扩展获取。
4. **RunTool 在非 running 状态返回 ToolResult 而非 error** — 保持与工具执行错误一致的模式（ToolResult.IsError = true）。
5. **状态转换并发安全** — 使用 `sync.RWMutex`：读状态用 RLock，写状态用 Lock。

## Alternatives Considered

1. **状态作为 string 类型** — 考虑过用 string 表示状态以便扩展。排除：用户确认使用固定状态集合，int iota 更类型安全且性能更好。
2. **channel-based 状态机** — 用 channel 序列化状态转换。排除：对于简单的状态字段读写，mutex 更直接且足够。
3. **在 core.Agent 接口中添加 Pause/Resume** — 排除：不修改 M1 已定义的接口，避免破坏性变更。
