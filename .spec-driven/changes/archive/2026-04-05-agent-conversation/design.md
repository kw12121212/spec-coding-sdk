# Design: agent-conversation

## Approach

在 `internal/agent/` 包中新增 `message.go` 和 `conversation.go`，定义消息类型和会话管理结构。通过 `BaseAgent` 扩展方法暴露会话操作，不修改 `core.Agent` 接口。

### 消息模型

定义 `Message` 结构体和角色常量，遵循与 claw-code 一致的语义：
- `RoleUser` — 用户输入
- `RoleAssistant` — 助手（LLM）响应
- `RoleTool` — 工具执行结果

每条消息包含 role、content、时间戳，以及可选的 tool name（仅 tool 角色消息）。

### 会话模型

`Conversation` 结构体管理有序消息列表，提供：
- `Add(msg)` — 追加消息并发出事件
- `Messages()` — 返回全部消息的快照（只读切片拷贝）
- `Len()` — 消息数量
- `Clear()` — 清空历史

`Conversation` 通过 `sync.RWMutex` 保证并发安全。

### Agent 集成

`BaseAgent` 新增：
- `Conversation() *Conversation` — 获取当前会话
- `SetConversation(c *Conversation)` — 设置会话（仅 StateInit 状态允许）

通过 `WithConversation(c *Conversation) Option` 在构造时注入。未注入时自动创建空会话。

## Key Decisions

1. **不修改 core.Agent 接口** — Conversation 作为 BaseAgent 扩展方法，保持接口最小化。后续接口层（JSON-RPC/HTTP/gRPC）按需暴露。
2. **Message 为值类型** — 消息创建后不可变，避免并发修改问题。
3. **Conversation 独立于 BaseAgent 状态** — 即使 agent 处于 stopped 状态，对话历史仍可读取，支持调试和重放。
4. **Messages() 返回快照** — 返回切片拷贝而非引用，防止外部修改内部状态。

## Alternatives Considered

1. **在 core 包定义 Conversation 接口** — 放弃，因为会话管理是 agent 的实现细节而非核心抽象，过早抽象增加不必要的接口复杂度。
2. **消息使用 interface{} 而非结构体** — 放弃，强类型 Message 更安全且便于 JSON 序列化。
3. **Conversation 内嵌到 BaseAgent 字段而非指针** — 放弃，指针注入允许外部预先构建会话、测试时替换，也避免 BaseAgent 结构体膨胀。
