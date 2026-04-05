# Tasks: agent-orchestrator

## Implementation

- [x] 定义 `ToolCall` 结构体（Name string + Input json.RawMessage）和 `ThinkResult` 结构体（Content string + ToolCalls []ToolCall）到 `internal/agent/orchestrator.go`
- [x] 定义 `Thinker` 接口（Think(ctx, []Message) (ThinkResult, error)）到 `internal/agent/orchestrator.go`
- [x] 定义 `ToolRegistry` 接口（Get(name string) (core.Tool, bool)）到 `internal/agent/orchestrator.go`
- [x] 定义编排循环事件类型常量（EventOrchestratorThink、EventOrchestratorAct、EventOrchestratorObserve、EventOrchestratorComplete）到 `internal/core/events.go`
- [x] 定义编排循环事件结构体（OrchestratorThinkEvent、OrchestratorActEvent、OrchestratorObserveEvent、OrchestratorCompleteEvent）到 `internal/core/events.go`
- [x] 实现 `Orchestrator` 结构体（持有 *BaseAgent、Thinker、ToolRegistry、maxIterations）和构造函数 `NewOrchestrator` 到 `internal/agent/orchestrator.go`
- [x] 实现 `WithMaxIterations(n int)` Option 到 `internal/agent/orchestrator.go`
- [x] 实现 `Orchestrator.Run(ctx, userMessage string) (ThinkResult, error)` 编排循环逻辑到 `internal/agent/orchestrator.go`

## Testing

- [x] 编写 mock Thinker 和 stub Tool 测试辅助到 `internal/agent/orchestrator_test.go`
- [x] 测试：单轮对话无工具调用（用户提问 → LLM 直接回复）
- [x] 测试：单轮工具调用（用户提问 → LLM 调用工具 → 工具结果 → LLM 最终回复）
- [x] 测试：多轮工具调用（连续 2+ 次工具调用后 LLM 给出最终回复）
- [x] 测试：工具调用失败时错误反馈给 LLM（LLM 决定重试或停止）
- [x] 测试：达到最大迭代次数时返回错误
- [x] 测试：Orchestrator 在 BaseAgent 非 running 状态下调用 Run 返回错误
- [x] 测试：编排循环发射正确的事件序列
- [x] 测试：WithMaxIterations Option 正确设置自定义迭代上限
- [x] Lint passes
- [x] Unit tests pass

## Verification

- [x] Verify implementation matches proposal scope
- [x] Verify all delta spec requirements are independently testable
