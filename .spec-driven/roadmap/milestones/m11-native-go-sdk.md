# M11 - Native Go SDK 层

## Goal

在 M1 核心接口之上封装面向第三方的公共 Go SDK facade，让外部可直接 import 使用 agent 的全部能力。

## In Scope

- 面向第三方的公共 SDK API（facade 层，封装 M1-M10 全部核心能力）
- 事件订阅与回调机制
- 统一错误处理模式
- SDK 使用文档和示例

## Out of Scope

- 非 Go 语言 SDK（未来可能）
- 具体传输层实现（M12-M14）
- 底层接口的重新定义（SDK 是 M1 接口的 facade，不替代它们）

## Done Criteria

- 第三方可通过 `go get` 引入 SDK 并创建 agent 实例
- SDK 事件回调可正确接收 agent 运行时事件
- 错误类型可被调用方正确判断和处理
- 有示例代码展示基本用法（创建 agent、注册工具、运行循环）

## Planned Changes

- `sdk-public-api` - 公共 SDK facade 接口定义与实现
- `sdk-events` - 事件订阅与回调机制实现
- `sdk-error-handling` - 统一错误类型与处理模式实现

## Dependencies

- M1-M5 核心能力（接口、工具集、agent 运行时、LLM 后端）
- M6 权限系统（SDK 需暴露权限配置）
- M9-M10 协议工具（LSP/MCP 作为可选工具注册）

## Risks

- SDK API 一旦发布很难做 breaking change，首次设计需充分评审
- 需平衡易用性与灵活性——facade 不应过度隐藏底层能力

## Status

- Declared: proposed

## Notes

- SDK 层是 M1 核心接口的 **facade**，不是替代。底层 M1 接口仍然可用，SDK 提供更高层次的便捷 API
- 参考 claw-code 的使用模式提炼常用操作为简洁的 SDK 调用
- M7-M8（注册表）为可选依赖，SDK 可延迟初始化注册表功能
