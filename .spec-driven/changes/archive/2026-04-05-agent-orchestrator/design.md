# Design: agent-orchestrator

## Approach

在 `internal/agent/` 包中新增编排相关类型和 `Orchestrator` 结构体。编排器作为 BaseAgent 之上的编排层，不修改 BaseAgent 本身。

核心流程：
1. 用户消息通过 `Orchestrator.Run(ctx, userMessage)` 触发
2. Run 将用户消息添加到 Conversation，进入循环
3. **Think**: 调用 `Thinker.Think(ctx, messages)` 获取 LLM 响应
4. 判断 ThinkResult 是否包含工具调用：
   - 无工具调用 → 将助手回复添加到 Conversation，返回结果
   - 有工具调用 → 进入 Act → Observe 步骤
5. **Act**: 通过 `ToolRegistry` 查找工具并执行，将工具结果作为 tool 消息添加到 Conversation
6. **Observe**: 回到步骤 3，携带更新的消息列表
7. 循环直到无工具调用或达到最大迭代次数

## Key Decisions

1. **Thinker 接口而非直接 LLM 调用**：编排器依赖 `Thinker` 接口而非具体 LLM 实现，使测试可用 mock，M5 可用真实实现替换。Thinker 接口定义在 `internal/agent/` 包中。

2. **ToolRegistry 接口用于工具查找**：编排器通过 `ToolRegistry` 接口按名称查找工具，而非直接持有工具列表。这使得工具注册方式（静态列表、动态注册）对编排器透明。

3. **无内置重试**：工具调用失败时，编排器将错误作为工具结果反馈给 Thinker，由 LLM 决定是否重试。

4. **默认最大迭代次数 50**：提供合理的默认值防止无限循环，同时通过 `WithMaxIterations(n int)` Option 允许覆盖。

5. **编排器与 BaseAgent 解耦**：Orchestrator 不嵌入 BaseAgent，而是持有 `*BaseAgent` 引用。BaseAgent 的生命周期由外部管理，Orchestrator 专注于编排逻辑。

6. **事件发射**：编排循环的每一步（think、act、observe）都通过 BaseAgent 的 EventEmitter 发射事件，便于外部观察和调试。

## Alternatives Considered

1. **将编排逻辑直接写入 BaseAgent**：会增加 BaseAgent 的复杂度，违反单一职责。选择独立的 Orchestrator 更清晰。

2. **使用 channel 驱动的状态机**：虽然更灵活，但当前循环逻辑足够简单，直接在 Run 方法中实现更直观。如果未来需要支持暂停/恢复中间步骤，可重构为 channel 模式。

3. **将 Thinker 接口放在 internal/core**：考虑过放在 core 包使其更通用，但当前只有编排器使用，放在 agent 包更符合最小暴露原则。M5 可通过实现该接口注入。
