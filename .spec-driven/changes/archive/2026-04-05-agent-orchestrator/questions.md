# Questions: agent-orchestrator

## Open

<!-- No open questions -->

## Resolved

- [x] Q: 编排器在工具调用失败时是否应有内置重试/退避行为？
  Context: 影响编排器复杂度和与 M5 LLM 后端的交互模式
  A: 不内置重试。编排器将错误作为工具结果反馈给 Thinker，由 LLM 决定是否重试。

- [x] Q: 循环最大迭代次数的默认值和可配置性？
  Context: 影响编排器的默认行为和 Option 设计
  A: 默认值 50，通过 WithMaxIterations Option 可配置。
