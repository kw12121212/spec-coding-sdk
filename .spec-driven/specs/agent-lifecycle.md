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

### Requirement: 消息类型定义

- 项目 MUST 在 `internal/agent/` 包中定义 `Role` 类型（基于 string），包含以下角色常量：`RoleUser`（"user"）、`RoleAssistant`（"assistant"）、`RoleTool`（"tool"）。
- 项目 MUST 在 `internal/agent/` 包中定义 `Message` 结构体，包含以下字段：
  - `Role` (Role) — 消息角色
  - `Content` (string) — 消息内容
  - `ToolName` (string) — 工具名称，仅当 Role 为 `RoleTool` 时有值，其他角色为零值
  - `Timestamp` (time.Time) — 消息创建时间
- `Message` MUST 通过 `NewMessage(role Role, content string) Message` 构造函数创建，自动设置 `Timestamp` 为当前时间。
- 当 `role` 为 `RoleTool` 时，MUST 通过 `NewToolMessage(toolName, content string) Message` 构造函数创建，同时设置 `ToolName`。

### Requirement: 会话消息管理

- 项目 MUST 在 `internal/agent/` 包中提供 `Conversation` 结构体，管理有序消息列表。
- `Conversation` MUST 通过 `NewConversation(opts ...ConversationOption) *Conversation` 构造函数创建。
- `Conversation` MUST 支持 `WithConversationEmitter(emitter core.EventEmitter) ConversationOption` 注入 `EventEmitter`。
- 当未注入 `EventEmitter` 时，`Conversation` MUST 正常工作（不发送事件，不报错）。
- `Add(msg Message)` MUST 将消息追加到内部列表，并通过 `EventEmitter`（如有）发出 `EventMessageAdded` 类型事件，payload 为 `MessageEvent`。
- `Messages() []Message` MUST 返回消息列表的快照（切片拷贝），对外部修改不影响内部状态。
- `Len() int` MUST 返回当前消息数量。
- `Clear()` MUST 清空所有消息，不发出事件。
- `Conversation` 的所有操作 MUST 是并发安全的（使用 `sync.RWMutex`）。

### Requirement: 会话事件类型

- 项目 MUST 在 `internal/core/events.go` 中新增 `EventMessageAdded = "message.added"` 事件类型常量。
- 项目 MUST 在 `internal/core/events.go` 中新增 `MessageEvent` 结构体：
  - `Role` (string) — 消息角色
  - `Content` (string) — 消息内容摘要（不超过 200 字符，超出截断加省略号）
  - `ToolName` (string) — 工具名称（可选）

### Requirement: Agent 会话集成

- `BaseAgent` MUST 新增 `conversation *Conversation` 字段。
- `BaseAgent` MUST 支持 `WithConversation(c *Conversation) Option` 在构造时注入会话。
- 当未注入 `Conversation` 时，`New()` MUST 自动创建空 `Conversation`。
- `BaseAgent` MUST 提供 `Conversation() *Conversation` 方法返回当前会话。
- `BaseAgent` MUST 提供 `SetConversation(c *Conversation) error` 方法替换会话，仅在 `StateInit` 状态下允许，其他状态 MUST 返回错误。

### Requirement: Thinker 接口定义

- 项目 MUST 在 `internal/agent/` 包中定义 `Thinker` 接口，包含方法 `Think(ctx context.Context, messages []Message) (ThinkResult, error)`。
- `Thinker` 接口 MUST 可由外部包实现（不依赖 `internal/agent/` 的非导出类型）。

### Requirement: ThinkResult 和 ToolCall 类型定义

- 项目 MUST 在 `internal/agent/` 包中定义 `ToolCall` 结构体，包含 `Name` (string) 和 `Input` (json.RawMessage) 字段。
- 项目 MUST 在 `internal/agent/` 包中定义 `ThinkResult` 结构体，包含 `Content` (string) 和 `ToolCalls` ([]ToolCall) 字段。
- `ThinkResult` 当 `ToolCalls` 为空切片或 nil 时，表示 LLM 给出最终文本回复（Content 有值）。
- `ThinkResult` 当 `ToolCalls` 非空时，表示 LLM 请求执行一个或多个工具。

### Requirement: ToolRegistry 接口定义

- 项目 MUST 在 `internal/agent/` 包中定义 `ToolRegistry` 接口，包含方法 `Get(name string) (core.Tool, bool)`。
- `ToolRegistry` 接口 MUST 可由外部包实现。

### Requirement: Orchestrator 结构体

- 项目 MUST 在 `internal/agent/` 包中提供 `Orchestrator` 结构体，持有 `*BaseAgent`、`Thinker`、`ToolRegistry` 和最大迭代次数配置。
- `Orchestrator` MUST 通过 `NewOrchestrator(agent *BaseAgent, thinker Thinker, registry ToolRegistry, opts ...OrchestratorOption) *Orchestrator` 构造函数创建。
- 项目 MUST 提供 `WithMaxIterations(n int) OrchestratorOption` 选项函数，允许覆盖默认最大迭代次数。
- 默认最大迭代次数 MUST 为 50。

### Requirement: Orchestrator Run 方法

- `Orchestrator` MUST 提供 `Run(ctx context.Context, userMessage string) (ThinkResult, error)` 方法执行完整的编排循环。
- `Run` MUST 在执行前检查 BaseAgent 是否处于 `StateRunning` 状态；若非 running 状态，MUST 返回错误。
- `Run` MUST 将用户消息（RoleUser）添加到 Conversation。
- 编排循环 MUST 遵循以下步骤：
  1. **Think**: 调用 `Thinker.Think(ctx, messages)` 获取 LLM 响应
  2. 若 ThinkResult 无工具调用 → 将助手消息（RoleAssistant）添加到 Conversation，返回 ThinkResult
  3. 若 ThinkResult 有工具调用 → 依次执行每个工具调用，将助手消息和工具结果消息添加到 Conversation
  4. **Observe**: 用更新后的消息列表回到步骤 1
- 每次循环迭代 MUST 检查迭代计数是否超过 maxIterations；超过时 MUST 返回错误。
- 每次循环迭代 MUST 检查 context 是否已取消；若已取消 MUST 立即返回 context 错误。
- 工具执行失败时，MUST 将错误信息作为 tool 消息（IsError 标记）添加到 Conversation，继续循环（不内置重试）。
- 工具在 registry 中未找到时，MUST 将 "tool not found" 错误作为 tool 消息添加到 Conversation，继续循环。

### Requirement: 编排事件类型

- 项目 MUST 在 `internal/core/events.go` 中新增以下事件类型常量：
  - `EventOrchestratorThink` = `"orchestrator.think"`
  - `EventOrchestratorAct` = `"orchestrator.act"`
  - `EventOrchestratorObserve` = `"orchestrator.observe"`
  - `EventOrchestratorComplete` = `"orchestrator.complete"`
- 项目 MUST 在 `internal/core/events.go` 中新增以下事件结构体：
  - `OrchestratorThinkEvent` — 包含 `Iteration` (int) 字段
  - `OrchestratorActEvent` — 包含 `Iteration` (int)、`ToolName` (string)、`Success` (bool) 字段
  - `OrchestratorObserveEvent` — 包含 `Iteration` (int)、`MessageCount` (int) 字段
  - `OrchestratorCompleteEvent` — 包含 `TotalIterations` (int)、`ToolCallsMade` (int) 字段
- 每种事件结构体 MUST 实现 `EventType() string` 方法返回对应的常量。
- 每种事件结构体 MUST 可 JSON 序列化/反序列化。

### Requirement: 编排循环事件发射

- Orchestrator MUST 在每次 Think 调用前通过 BaseAgent 的 EventEmitter 发射 `OrchestratorThinkEvent`。
- Orchestrator MUST 在每次工具执行后发射 `OrchestratorActEvent`（Success 根据工具结果设置）。
- Orchestrator MUST 在工具结果添加到 Conversation 后发射 `OrchestratorObserveEvent`。
- Orchestrator MUST 在循环正常结束时（无工具调用或最终回复）发射 `OrchestratorCompleteEvent`。
- 当 BaseAgent 的 EventEmitter 为 nil 时，MUST 不发射事件且不报错。
