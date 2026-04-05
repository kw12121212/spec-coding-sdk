# Tasks: agent-conversation

## Implementation

- [x] 在 `internal/agent/message.go` 中定义 `Role` 类型和常量（`RoleUser`、`RoleAssistant`、`RoleTool`）
- [x] 在 `internal/agent/message.go` 中定义 `Message` 结构体（Role、Content、Timestamp、ToolName 字段）及 `NewMessage(role, content)` 构造函数
- [x] 在 `internal/agent/conversation.go` 中定义 `Conversation` 结构体（messages 切片、emitter、mutex）
- [x] 实现 `Conversation.Add(msg Message)` 方法：追加消息、发出 `message.added` 事件
- [x] 实现 `Conversation.Messages() []Message` 方法：返回消息快照
- [x] 实现 `Conversation.Len() int` 方法
- [x] 实现 `Conversation.Clear()` 方法
- [x] 在 `internal/core/events.go` 中新增 `EventMessageAdded` 常量和 `MessageEvent` 结构体
- [x] 在 `internal/agent/agent.go` 中新增 `conversation` 字段和 `WithConversation` Option
- [x] 在 `internal/agent/agent.go` 中新增 `Conversation() *Conversation` 和 `SetConversation(c *Conversation)` 方法
- [x] 确保 `NewBaseAgent` 未注入 Conversation 时自动创建空会话

## Testing

- [x] `message_test.go`：验证 Role 常量值、Message 构造、字段正确性
- [x] `conversation_test.go`：验证 Add 追加消息、Messages 返回快照（修改快照不影响内部）、Len 正确、Clear 清空、并发安全
- [x] `conversation_test.go`：验证有/无 EventEmitter 时 Add 的行为
- [x] `agent_test.go`：验证 WithConversation 注入、默认空会话、SetConversation 仅 StateInit 可用
- [x] `go vet` 和 lint 通过

## Verification

- [x] 所有测试通过 (`go test ./internal/agent/... ./internal/core/...`)
- [x] Delta spec 反映实际实现
