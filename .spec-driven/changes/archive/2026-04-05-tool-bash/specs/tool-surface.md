# tool-surface

## ADDED Requirements

### Requirement: Bash 命令执行工具

- 项目 MUST 在 `internal/tools/bash/` 包中提供 `Tool` 结构体，实现 `internal/core` 包的 `Tool` 接口。
- `Tool.Execute` MUST 接受 JSON 输入（`Input` 结构体），包含以下字段：
  - `command` (string, 必填) — 要执行的 bash 命令
  - `timeout` (int, 可选) — 超时秒数，默认 120
  - `working_dir` (string, 可选) — 工作目录，默认继承进程工作目录
- 当 `command` 为空字符串或输入 JSON 无法解析时，`Execute` MUST 返回非 nil error。
- `Tool` MUST 通过 `/bin/bash -c` 执行命令，合并捕获 stdout 和 stderr 到单一输出。
- `Tool.Execute` MUST 支持超时控制：当执行时间超过指定的 timeout 秒数时，MUST 终止子进程并返回 `ToolResult`，其中 `IsError` 为 true，`Output` 包含超时指示信息。
- `Tool` MUST 限制单次执行输出大小为 1MB（1048576 字节）。超出时 MUST 截断输出并在 `Output` 末尾附加截断提示，同时设置 `IsError` 为 true。
- 命令以非零退出码结束时，`ToolResult.IsError` MUST 为 true。
- 命令以零退出码结束时，`ToolResult.IsError` MUST 为 false，`ToolResult.Output` 包含合并的 stdout+stderr 输出。

### Requirement: Bash 工具权限检查钩子

- `Tool` MUST 支持可选的 `PermissionProvider` 注入（通过 `New(perms)` 构造函数）。
- 当 `PermissionProvider` 非 nil 时，`Execute` MUST 在执行命令前调用 `Check(ctx, "bash:execute", command)`。
- 当 `PermissionProvider.Check` 返回非 nil error 时，`Execute` MUST 不执行命令，直接返回 `ToolResult{IsError: true, Output: error message}`。
- 当 `PermissionProvider` 为 nil 时，`Execute` MUST 跳过权限检查，直接执行命令。
