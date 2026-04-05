# agent-lifecycle

## ADDED Requirements

### Requirement: Agent 状态类型定义

- 项目 MUST 在 `internal/agent/` 包中定义 `State` 类型（基于 int），包含以下固定状态常量：`StateInit`、`StateRunning`、`StatePaused`、`StateStopped`、`StateError`。
- 每个状态常量 MUST 有导出注释说明其含义。
- `State` MUST 实现 `String() string` 方法，返回状态的小写字符串表示（"init"、"running"、"paused"、"stopped"、"error"）。
- 项目 MUST 在 `internal/agent/` 包中定义合法状态转换表，仅允许以下转换：
  - init → running（通过 Start）
  - running → paused（通过 Pause）
  - paused → running（通过 Resume）
  - running → stopped（通过 Stop）
  - paused → stopped（通过 Stop）
  - * → error（内部错误触发）
  - error → stopped（通过 Stop）
- 非法状态转换 MUST 返回错误。

### Requirement: BaseAgent 结构体

- 项目 MUST 在 `internal/agent/` 包中提供 `BaseAgent` 结构体，实现 `internal/core.Agent` 接口（Start、Stop、RunTool 方法）。
- `BaseAgent` MUST 通过 `New(opts ...Option) *BaseAgent` 构造函数创建，支持通过 functional options 注入可选依赖。
- `BaseAgent` MUST 支持通过 `WithEmitter(emitter core.EventEmitter) Option` 注入 `EventEmitter`。
- 当未注入 `EventEmitter` 时，`BaseAgent` MUST 正常工作（不发送事件，不报错）。
- `BaseAgent` 初始状态 MUST 为 `StateInit`。
- 项目 MUST 包含编译期接口满足检查 `var _ core.Agent = (*BaseAgent)(nil)`。

### Requirement: Agent 生命周期方法

- `Start(_ context.Context)` MUST 将状态从 `StateInit` 转换为 `StateRunning`，并通过 `EventEmitter`（如有）发出 `AgentStateEvent`（state="running"）。
- `Start` 在非 `StateInit` 状态下调用 MUST 返回错误。
- `Stop(_ context.Context)` MUST 将状态转换为 `StateStopped`（仅允许从 running、paused、error 状态调用），并通过 `EventEmitter`（如有）发出 `AgentStateEvent`（state="stopped"）。
- `Stop` 在 `StateInit` 或 `StateStopped` 状态下调用 MUST 返回错误。
- `Pause(_ context.Context)` MUST 将状态从 `StateRunning` 转换为 `StatePaused`，并通过 `EventEmitter`（如有）发出 `AgentStateEvent`（state="paused"）。此方法不在 `core.Agent` 接口中，为 `BaseAgent` 的扩展方法。
- `Pause` 在非 `StateRunning` 状态下调用 MUST 返回错误。
- `Resume(_ context.Context)` MUST 将状态从 `StatePaused` 转换为 `StateRunning`，并通过 `EventEmitter`（如有）发出 `AgentStateEvent`（state="running"）。此方法不在 `core.Agent` 接口中，为 `BaseAgent` 的扩展方法。
- `Resume` 在非 `StatePaused` 状态下调用 MUST 返回错误。

### Requirement: Agent RunTool 方法

- `RunTool(ctx context.Context, tool core.Tool, input json.RawMessage)` MUST 仅在 `StateRunning` 状态下执行 `tool.Execute(ctx, input)` 并返回其结果。
- 在非 `StateRunning` 状态下调用 `RunTool` MUST 返回 `(core.ToolResult{IsError: true}, nil)`，`Output` 包含当前状态和"not running"提示信息。

### Requirement: Agent 状态查询

- `BaseAgent` MUST 提供 `State() State` 方法返回当前状态。
- 状态读取 MUST 是并发安全的。

### Requirement: Agent 状态转换并发安全

- `BaseAgent` 的所有状态变更操作 MUST 是并发安全的（使用 `sync.RWMutex`）。
- 多个 goroutine 同时调用 Start/Stop/Pause/Resume MUST 不会导致数据竞争或无效状态。
