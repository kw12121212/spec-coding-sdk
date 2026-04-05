# tool-surface

## ADDED Requirements

### Requirement: 文件读取工具

- 项目 MUST 在 `internal/tools/fileops/` 包中提供 `ReadTool` 结构体，实现 `internal/core` 包的 `Tool` 接口。
- `ReadTool.Execute` MUST 接受 JSON 输入（`ReadInput` 结构体），包含以下字段：
  - `file_path` (string, 必填) — 要读取的文件绝对路径
  - `offset` (int, 可选) — 起始行号（0-based），默认 0（从文件开头）
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
