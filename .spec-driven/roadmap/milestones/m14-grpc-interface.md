# M14 - gRPC 接口

## Goal

实现基于 Protocol Buffers 的 gRPC 接口层，提供高性能、类型安全的远程调用能力。

## In Scope

- .proto 文件定义与 Go 代码生成
- gRPC 服务实现（映射到 SDK 操作）
- 流式响应支持（agent 事件流、工具执行输出流）
- gRPC 层的端到端测试

## Out of Scope

- gRPC-Web / gRPC-Gateway（如需要可纳入后续迭代）
- 跨语言客户端生成（仅生成 Go stub）

## Done Criteria

- gRPC 客户端可成功调用所有 agent 操作
- 流式 RPC 可正确推送 agent 运行时事件
- .proto 文件可通过 `protoc` 正确生成 Go 代码
- 有端到端测试覆盖 gRPC 请求-响应和流式推送

## Planned Changes

- `grpc-proto` - Declared: planned - .proto 定义文件与代码生成配置
- `grpc-service` - Declared: planned - gRPC 服务端实现
- `grpc-streaming` - Declared: planned - 流式 RPC 实现（事件流、输出流）
- `grpc-e2e-tests` - Declared: planned - gRPC 层端到端测试实现

## Dependencies

- M11 Native Go SDK 层
- Protocol Buffers 编译工具链

## Risks

- Protobuf schema 的向后兼容性约束
- 流式 RPC 的连接管理和错误恢复

## Status

- Declared: proposed

## Notes

- .proto 定义需考虑多语言客户端兼容性，即使首期只生成 Go stub
- 参考 claw-code gRPC 接口的行为定义
- 与 M12、M13 可并行开发
