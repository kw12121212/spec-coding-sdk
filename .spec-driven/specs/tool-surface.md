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

### Requirement: 文件读取工具

- 项目 MUST 在 `internal/tools/fileops/` 包中提供 `ReadTool` 结构体，实现 `internal/core` 包的 `Tool` 接口。
- `ReadTool.Execute` MUST 接受 JSON 输入（`ReadInput` 结构体），包含以下字段：
  - `file_path` (string, 必填) — 要读取的文件绝对路径
  - `offset` (int, 可选) — 起始行号（1-based），默认 0（从文件开头）
  - `limit` (int, 可选) — 最大返回行数，0 表示不限制
- 当 `file_path` 为空、不是绝对路径或输入 JSON 无法解析时，`Execute` MUST 返回非 nil error。
- 当文件不存在或无法读取时，`Execute` MUST 返回 `ToolResult{IsError: true}`，`Output` 包含错误描述。
- `ReadTool.Execute` MUST 返回文件内容，每行前缀为行号和制表符（格式：`<行号>\t<行内容>`），行号从 1 开始。
- 当指定 `offset` 时，MUST 从第 offset 行开始返回（跳过前 offset 行）。
- 当指定 `limit` 且 limit > 0 时，MUST 最多返回 limit 行内容。
- 当文件不存在或路径指向目录时，MUST 返回错误结果。

### Requirement: 文件写入工具

- 项目 MUST 在 `internal/tools/fileops/` 包中提供 `WriteTool` 结构体，实现 `internal/core` 包的 `Tool` 接口。
- `WriteTool.Execute` MUST 接受 JSON 输入（`WriteInput` 结构体），包含以下字段：
  - `file_path` (string, 必填) — 要写入的文件绝对路径
  - `content` (string, 必填) — 要写入的文件内容
- 当 `file_path` 为空、不是绝对路径或输入 JSON 无法解析时，`Execute` MUST 返回非 nil error。
- 当 `content` 为空字符串时，`Execute` MUST 正常写入空文件（不视为错误）。
- `WriteTool.Execute` MUST 在写入前自动创建所有不存在的父目录。
- 写入成功时，MUST 返回 `ToolResult{IsError: false}`。
- 当父目录创建失败或写入失败时，MUST 返回 `ToolResult{IsError: true}`。

### Requirement: 文件编辑工具

- 项目 MUST 在 `internal/tools/fileops/` 包中提供 `EditTool` 结构体，实现 `internal/core` 包的 `Tool` 接口。
- `EditTool.Execute` MUST 接受 JSON 输入（`EditInput` 结构体），包含以下字段：
  - `file_path` (string, 必填) — 要编辑的文件绝对路径
  - `old_string` (string, 必填) — 要替换的原始字符串
  - `new_string` (string, 必填) — 替换后的新字符串
  - `replace_all` (bool, 可选, 默认 false) — 是否替换所有匹配
- 当 `file_path` 为空、不是绝对路径、`old_string` 为空或输入 JSON 无法解析时，`Execute` MUST 返回非 nil error。
- 当 `old_string` 在文件中未找到时，`Execute` MUST 返回 `ToolResult{IsError: true}`，`Output` 包含未找到的提示。
- 当 `replace_all` 为 false 且 `old_string` 在文件中出现多次时，`Execute` MUST 返回 `ToolResult{IsError: true}`，`Output` 包含多处匹配的提示。MUST NOT 执行任何替换。
- 当 `replace_all` 为 true 时，`Execute` MUST 替换所有匹配的 `old_string`。
- 替换成功时，`Output` MUST 包含替换次数信息。
- 当文件不存在时，MUST 返回错误结果。

### Requirement: 文件操作工具权限检查钩子

- `ReadTool`、`WriteTool`、`EditTool` MUST 各自支持可选的 `PermissionProvider` 注入（通过 `NewReadTool(perms)`、`NewWriteTool(perms)`、`NewEditTool(perms)` 构造函数）。
- 当 `PermissionProvider` 非 nil 时，各工具 MUST 在执行操作前调用 `Check`：
  - `ReadTool`：`Check(ctx, "file:read", file_path)`
  - `WriteTool`：`Check(ctx, "file:write", file_path)`
  - `EditTool`：`Check(ctx, "file:edit", file_path)`
- 当 `PermissionProvider.Check` 返回非 nil error 时，`Execute` MUST 不执行操作，直接返回 `ToolResult{IsError: true, Output: error message}`。
- 当 `PermissionProvider` 为 nil 时，`Execute` MUST 跳过权限检查，直接执行操作。

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
