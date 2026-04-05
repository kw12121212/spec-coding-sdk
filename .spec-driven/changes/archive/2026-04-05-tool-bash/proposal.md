# tool-bash

## What

实现 Bash 命令执行工具，作为第一个符合 M1 `Tool` 接口的具体工具实现。支持命令运行、超时控制和 stdout/stderr 输出捕获。

## Why

- M1 已定义 `Tool` 接口（`Execute(ctx, json.RawMessage) (ToolResult, error)`），需要第一个具体实现来验证接口设计的完备性
- Bash 执行是 coding agent 最核心的能力，是 agent 调用其他工具的基础
- 为后续 tool-file-ops、tool-grep、tool-glob 建立工具实现的参考模式

## Scope

In scope:
- `BashTool` 结构体实现 `Tool` 接口
- 命令输入 schema（command、timeout、working_dir）
- 命令执行（通过 `os/exec`）并捕获 stdout + stderr
- 超时控制（基于 `context.WithTimeout`）
- 输出大小限制（防止内存溢出）
- 权限检查钩子（接收 `PermissionProvider`，执行前调用 `Check`）
- 单元测试覆盖 happy path、timeout、output-too-large、permission-denied

Out of scope:
- 交互式命令支持（如 `gcloud auth login`）
- 命令注入防护策略的具体规则（M5 权限模型）
- 并发执行的资源隔离（cgroups 等）
- shell 环境变量自定义（后续 MAY 扩展）

## Unchanged Behavior

- `internal/core/` 中的 `Tool`、`ToolResult`、`PermissionProvider` 接口不变
- `pkg/sdk/` 和 `api/proto/` 包不变
