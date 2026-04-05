# Questions: core-interfaces

## Open

<!-- No open questions -->

## Resolved

- [x] Q: 核心接口应定义在 `internal/core/` 还是 `pkg/sdk/`？
  Context: 决定接口的可见性和 M10 的职责边界
  A: 定义在 `internal/core/`。M10 的目标就是封装公共 SDK facade，现在放入 `pkg/` 会抢占 M10 范围。

- [x] Q: `Tool` 接口是否需要单独的 `Validate` 方法？
  Context: 影响工具合约的复杂度和后续所有工具实现
  A: 不需要。claw-code 的工具表面中并非所有工具都需要独立校验，单方法 `Execute` 满足当前已知需求。如后续需要，可通过接口组合添加。
