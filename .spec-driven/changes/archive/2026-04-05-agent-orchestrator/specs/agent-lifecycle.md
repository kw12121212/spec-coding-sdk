# agent-lifecycle

## ADDED Requirements

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
