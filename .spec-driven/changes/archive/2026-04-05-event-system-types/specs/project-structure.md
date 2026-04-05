# project-structure

## ADDED Requirements

### Requirement: 事件系统类型

- 项目 MUST 在 `internal/core/` 包中定义以下具体事件类型：
  - `ToolCallEvent` — 工具调用事件，MUST 包含工具名称（`ToolName`）和输入参数（`Input`）字段
  - `ToolResultEvent` — 工具结果事件，MUST 包含工具名称（`ToolName`）和执行结果（`Result`）字段
  - `AgentStateEvent` — Agent 状态变化事件，MUST 包含状态标识（`State`）和描述（`Message`）字段
  - `ErrorEvent` — 错误事件，MUST 包含错误码（`Code`）和错误信息（`Message`）字段
- 每种具体事件类型 MUST 可序列化为 JSON 并可从 JSON 反序列化（round-trip 一致）
- 项目 MUST 定义事件类型常量字符串，每种事件类型对应一个唯一常量
- 项目 MUST 在 `internal/core/` 包中定义 `EventEmitter` 接口，包含 `Emit(event Event)` 方法
- 项目 MUST 在 `internal/core/` 包中定义 `EventSubscriber` 接口，包含 `Subscribe(eventType string, handler func(Event))` 方法
- `EventEmitter` 和 `EventSubscriber` 接口 MUST 可由外部包实现
- `pkg/sdk/` 和 `api/proto/` 包 MUST NOT 被此变更修改
