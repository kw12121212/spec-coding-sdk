# project-structure

## ADDED Requirements

### Requirement: 核心接口定义

- 项目 MUST 在 `internal/core/` 包中定义以下核心接口：
  - `Tool` — 工具调用合约，MUST 包含 `Execute` 方法
  - `Agent` — Agent 生命周期合约，MUST 支持会话管理（启动、停止）和工具调用
  - `PermissionProvider` — 权限检查合约，MUST 包含 `Check` 方法
- 项目 MUST 在 `internal/core/` 包中定义 `Event` 结构体，包含类型标识、负载和时间戳字段
- 项目 MUST 定义 `ToolResult` 类型，作为 `Tool.Execute` 的返回值
- 所有核心类型 MUST 可被其他包引用（`go build ./...` 通过）
- `pkg/sdk/` 和 `api/proto/` 包 MUST NOT 被此变更修改

### Requirement: 接口可实现性

- 每个核心接口 MUST 可由外部包的类型实现（不依赖 `internal/core/` 的非导出类型）
- 测试 MUST 验证其他包可以定义实现这些接口的具体类型
