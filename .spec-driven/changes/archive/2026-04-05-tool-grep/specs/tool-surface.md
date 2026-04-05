# tool-surface

## ADDED Requirements

### Requirement: Grep 内容搜索工具

- 项目 MUST 在 `internal/tools/grep/` 包中提供 `Tool` 结构体，实现 `internal/core` 包的 `Tool` 接口。
- `Tool.Execute` MUST 接受 JSON 输入（`Input` 结构体），包含以下字段：
  - `pattern` (string, 必填) — 要搜索的正则表达式模式
  - `path` (string, 可选) — 搜索的文件或目录路径，默认为当前工作目录
  - `glob` (string, 可选) — 文件名过滤模式（如 `*.go`），传递给 rg 的 `--glob` 参数
  - `type` (string, 可选) — 文件类型过滤（如 `go`、`python`），传递给 rg 的 `--type` 参数
  - `output_mode` (string, 可选) — 输出模式，取值为 `content`（默认）、`files_with_matches`、`count`
  - `ignore_case` (bool, 可选, 默认 false) — 是否忽略大小写，传递给 rg 的 `-i` 参数
  - `context` (int, 可选) — 上下文行数，传递给 rg 的 `-C` 参数
  - `head_limit` (int, 可选) — 最大输出行数，0 表示不限制
- 当 `pattern` 为空或输入 JSON 无法解析时，`Execute` MUST 返回非 nil error。
- `Tool` MUST 通过 `exec.Command` 调用 `rg` 命令执行搜索，将 `pattern` 和 `path` 作为位置参数。
- `output_mode` MUST 映射为对应的 rg 参数：`content` → 无额外参数（默认行为），`files_with_matches` → `-l`，`count` → `-c`。
- `Tool.Execute` MUST 合并捕获 stdout 和 stderr 到单一输出。
- `Tool.Execute` MUST 限制单次执行输出大小为 1MB（1048576 字节）。超出时 MUST 截断输出并在 `Output` 末尾附加截断提示，同时设置 `IsError` 为 true。
- 当 `rg` 命令以退出码 0 结束时（找到匹配），`ToolResult.IsError` MUST 为 false。
- 当 `rg` 命令以退出码 1 结束时（未找到匹配），`ToolResult.IsError` MUST 为 false，`ToolResult.Output` MUST 包含无匹配的提示信息。
- 当 `rg` 命令以退出码 2+ 结束时（错误），`ToolResult.IsError` MUST 为 true。
- 当 `rg` 命令在 PATH 中不存在时，`Execute` MUST 返回 `ToolResult{IsError: true}`，`Output` 包含 rg 未找到的错误信息。

### Requirement: Grep 工具权限检查钩子

- `Tool` MUST 支持可选的 `PermissionProvider` 注入（通过 `New(perms)` 构造函数）。
- 当 `PermissionProvider` 非 nil 时，`Execute` MUST 在执行搜索前调用 `Check(ctx, "grep:execute", pattern)`。
- 当 `PermissionProvider.Check` 返回非 nil error 时，`Execute` MUST 不执行搜索，直接返回 `ToolResult{IsError: true, Output: error message}`。
- 当 `PermissionProvider` 为 nil 时，`Execute` MUST 跳过权限检查，直接执行搜索。
