# agent-lifecycle delta

## ADDED Requirements

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
