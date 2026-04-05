# Tasks: agent-lifecycle

## Implementation

- [x] 在 `internal/agent/` 包中定义 `AgentState` 类型和五个状态常量（StateInit, StateRunning, StatePaused, StateStopped, StateError）
- [x] 定义合法状态转换表（map 或函数），验证转换合法性
- [x] 实现 `BaseAgent` 结构体，包含 state 字段、emitter 字段和 sync.RWMutex
- [x] 实现 `Option` 类型和 `WithEmitter` functional option
- [x] 实现 `New(opts ...Option) *BaseAgent` 构造函数
- [x] 实现 `Start(ctx)` 方法（init → running 转换 + 事件发送）
- [x] 实现 `Stop(ctx)` 方法（running/paused/error → stopped 转换 + 事件发送）
- [x] 实现 `Pause(ctx)` 方法（running → paused 转换 + 事件发送）
- [x] 实现 `Resume(ctx)` 方法（paused → running 转换 + 事件发送）
- [x] 实现 `RunTool(ctx, tool, input)` 方法（running 状态检查 + 委托执行）
- [x] 实现 `State()` 方法（并发安全读取当前状态）
- [x] 验证 `BaseAgent` 满足 `core.Agent` 接口（编译期检查）

## Testing

- [x] 测试新建 agent 初始状态为 StateInit
- [x] 测试 Start 成功转换（init → running）及事件发送
- [x] 测试 Start 在非法状态下返回错误（已 running、已 stopped 等）
- [x] 测试 Stop 成功转换（running → stopped、paused → stopped、error → stopped）及事件发送
- [x] 测试 Stop 在非法状态下返回错误（init、已 stopped）
- [x] 测试 Pause 成功转换（running → paused）及事件发送
- [x] 测试 Pause 在非 running 状态返回错误
- [x] 测试 Resume 成功转换（paused → running）及事件发送
- [x] 测试 Resume 在非 paused 状态返回错误
- [x] 测试 RunTool 在 running 状态正确委托执行并返回工具结果
- [x] 测试 RunTool 在非 running 状态返回 IsError=true 的 ToolResult
- [x] 测试无 EventEmitter 时不报错、正常工作
- [x] 测试 EventEmitter 收到正确的 AgentStateEvent（state 和 message 字段）
- [x] 并发安全测试：多个 goroutine 同时调用生命周期方法不产生 data race
- [x] `make lint` 通过
- [x] `make test` 通过

## Verification

- [x] 验证 BaseAgent 满足 core.Agent 接口
- [x] 验证状态转换规则与 spec 一致
- [x] 验证所有测试独立运行（无共享可变状态）
