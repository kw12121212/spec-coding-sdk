# tool-gh

## What

实现 GitHub CLI 集成工具，新增一个符合 M1 `Tool` 接口的 `gh` 执行工具。
该工具接收结构化 JSON 输入，将参数列表转发给 `gh` 命令，使用 M3
已完成的内置工具管理器解析可执行文件路径，并返回合并后的命令输出。

## Why

- M3 当前只剩 `tool-gh` 与 `tool-rtk` 两项未完成，`tool-gh` 是直接消费
  已落地 `builtin-tool-manager` 的最小闭环
- 现有 roadmap 已要求支持 `gh` 作为 agent 可调用的外部工具，但主 specs
  还没有描述其可观察行为
- 该变更可建立“外部 CLI + 内置工具管理器”模式，为后续 `tool-rtk`
  提供更清晰的参照

## Scope

In scope:
- 在 `internal/tools/gh/` 中定义 `Tool` 及其 JSON 输入 schema
- 通过内置工具管理器解析 `gh` 可执行文件，优先使用宿主安装，必要时回退到
  SDK 托管安装
- 执行 `gh` 子进程，支持 `args`、`working_dir`、`timeout`
- 合并捕获 stdout 和 stderr，并沿用现有工具的输出大小限制约定
- 提供可选的权限检查钩子
- 为成功执行、托管安装回退、无效输入、权限拒绝、超时和命令失败补充测试

Out of scope:
- 对 GitHub API 做更高层抽象（如单独建模 issue/PR/repo 操作）
- 交互式认证流程（如 `gh auth login` 的 TTY 支持）
- 单次调用级别的环境变量覆盖
- `tool-rtk`、LSP/MCP、SDK/JSON-RPC/HTTP/gRPC 接口层变更

## Unchanged Behavior

Behaviors that must not change as a result of this change (leave blank if nothing is at risk):
- `builtin-tool-manager` 在 [builtin-tools] 中已定义的解析、下载和安装契约不变
- 现有 `tool-bash`、`tool-file-ops`、`tool-grep`、`tool-glob` 的行为不变
- 权限模型与 agent 生命周期行为不在本变更中扩展
