# tool-file-ops

## What

实现文件读/写/编辑三个工具，作为 M2 Tool Surface 基础工具集的第二组工具。提供 `ReadTool`、`WriteTool`、`EditTool` 三个结构体，均实现 M1 定义的 `Tool` 接口。

## Why

- 文件操作是 coding agent 的第二大核心能力（仅次于 bash 执行），agent 需要读取、创建和修改文件来完成代码编写任务
- `tool-bash` 已验证 `Tool` 接口的完备性，本变更继续在相同模式下扩展工具集
- `EditTool` 的精确字符串替换是 agent 精确修改代码的关键能力，避免全文覆写带来的 diff 噪音

## Scope

In scope:
- `ReadTool` — 读取文件内容，支持行号偏移和行数限制
- `WriteTool` — 创建或覆盖文件，支持自动创建父目录
- `EditTool` — 精确字符串替换，支持单次替换和全局替换
- 三个工具均支持可选的 `PermissionProvider` 权限检查钩子
- `file_path` 必须为绝对路径的验证
- 单元测试覆盖 happy path、error case、边界情况

Out of scope:
- 文件权限/模式管理（chmod 等）
- 文件锁和并发安全（当前阶段每个工具调用独立）
- 二进制文件读写（仅处理文本文件）
- 符号链接解析策略的深度定制
- 目录列表/遍历（属于 glob 工具范畴）

## Unchanged Behavior

- `internal/core/` 中的 `Tool`、`ToolResult`、`PermissionProvider` 接口不变
- `internal/tools/bash/` 包不变
- `pkg/sdk/` 和 `api/proto/` 包不变
