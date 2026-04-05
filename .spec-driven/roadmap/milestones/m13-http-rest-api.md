# M13 - HTTP REST API

## Goal

实现 HTTP REST API 接口层，提供网络可达的 agent 服务。

## In Scope

- API 路由定义（agent 操作、工具调用、状态查询）
- 认证中间件（API key / token）
- 请求/响应模型（JSON 序列化）
- 基础限流中间件
- HTTP 层的端到端测试

## Out of Scope

- WebSocket 长连接（如需要可纳入后续迭代）
- 完整的 OAuth2 / OIDC 实现

## Done Criteria

- 所有 API 端点可通过 HTTP 客户端正常调用
- 未认证请求被正确拒绝
- 限流在超阈值时返回 429
- 有端到端测试覆盖核心端点（从 HTTP 请求到 agent 操作再到 HTTP 响应）

## Planned Changes

- `http-routes` - Declared: planned - REST API 路由定义与 handler 实现
- `http-middleware` - Declared: planned - 认证与限流中间件实现
- `http-models` - Declared: planned - 请求/响应 JSON 模型定义
- `http-e2e-tests` - Declared: planned - HTTP 层端到端测试实现

## Dependencies

- M11 Native Go SDK 层
- M4 Agent 生命周期（长运行任务的状态查询）
- M6 权限系统（认证集成）

## Risks

- HTTP 接口的长时间运行任务需要合理的超时和异步响应策略
- API 版本管理策略需要提前确定

## Status

- Declared: proposed

## Notes

- REST API 设计参考 claw-code 现有的 HTTP 接口模式
- 使用 Go 标准库或轻量框架，避免重型依赖
- 与 M12、M14 可并行开发
