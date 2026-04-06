# tool-surface

## ADDED Requirements

### Requirement: GitHub CLI 集成工具

- 项目 MUST 在 `internal/tools/gh/` 包中提供 `Tool` 结构体，实现
  `internal/core` 包的 `Tool` 接口。
- `Tool.Execute` MUST 接受 JSON 输入（`Input` 结构体），包含以下字段：
  - `args` ([]string, 必填) — 传递给 `gh` 命令的参数列表
  - `working_dir` (string, 可选) — 工作目录，默认继承进程工作目录
  - `timeout` (int, 可选) — 超时秒数，默认 120
- 当 `args` 为空、任一参数为空字符串、或输入 JSON 无法解析时，
  `Execute` MUST 返回非 nil error。
- `Tool` MUST 在执行前解析一个可执行的 `gh` 文件路径，并使用解析出的
  可执行文件配合 `args` 启动子进程。
- 当宿主 PATH 中不存在 `gh`，但 SDK 托管安装可用时，`Tool.Execute`
  MUST 仍然能够执行该命令。
- `Tool.Execute` MUST 继承当前进程环境变量，且 MUST NOT 要求调用方为单次
  调用额外传入环境变量覆盖。
- `Tool.Execute` MUST 合并捕获 stdout 和 stderr 到单一输出。
- `Tool.Execute` MUST 支持超时控制：当执行时间超过指定 timeout 秒数时，
  MUST 终止子进程并返回 `ToolResult`，其中 `IsError` 为 true，`Output`
  包含超时指示信息。
- `Tool.Execute` MUST 限制单次执行输出大小为 1MB（1048576 字节）。超出时
  MUST 截断输出并在 `Output` 末尾附加截断提示，同时设置 `IsError` 为
  true。
- 当 `gh` 命令以零退出码结束时，`ToolResult.IsError` MUST 为 false。
- 当 `gh` 命令以非零退出码结束时，`ToolResult.IsError` MUST 为 true。
- 当 `gh` 可执行文件无法解析、下载、安装或启动时，`Execute` MUST 返回
  `ToolResult{IsError: true}`，`Output` 包含错误描述。

### Requirement: GitHub CLI 工具权限检查钩子

- `Tool` MUST 支持可选的 `PermissionProvider` 注入。
- 当 `PermissionProvider` 非 nil 时，`Execute` MUST 在解析或执行 `gh`
  前调用 `Check(ctx, "gh:execute", strings.Join(args, " "))`。
- 当 `PermissionProvider.Check` 返回非 nil error 时，`Execute` MUST 不执行
  命令，直接返回 `ToolResult{IsError: true, Output: error message}`。
- 当 `PermissionProvider` 为 nil 时，`Execute` MUST 跳过权限检查，直接执行
  命令。
