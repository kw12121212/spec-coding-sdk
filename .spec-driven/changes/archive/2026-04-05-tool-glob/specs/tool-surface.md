# tool-surface

## ADDED Requirements

### Requirement: 文件模式匹配工具

- 项目 MUST 在 `internal/tools/glob/` 包中提供 `Tool` 结构体，实现 `internal/core` 包的 `Tool` 接口。
- `Tool.Execute` MUST 接受 JSON 输入（`Input` 结构体），包含以下字段：
  - `pattern` (string, 必填) — glob 模式（支持 `*`、`?`、`**` 等）
  - `path` (string, 可选) — 搜索的目录路径，默认为当前工作目录
- 当 `pattern` 为空或输入 JSON 无法解析时，`Execute` MUST 返回非 nil error。
- `Tool` MUST 使用 Go 标准库（`filepath.WalkDir`）进行目录遍历，不依赖外部命令。
- `Tool` MUST 支持 `**` 模式用于递归匹配任意层级的目录。
- `Tool.Execute` MUST 返回匹配的文件路径，每行一个路径，按文件修改时间排序（最近的优先）。
- `Tool.Execute` MUST 限制单次执行输出大小为 1MB（1048576 字节）。超出时 MUST 截断输出并在 `Output` 末尾附加截断提示，同时设置 `IsError` 为 true。
- 当指定的 `path` 不存在或不是目录时，`Execute` MUST 返回 `ToolResult{IsError: true}`，`Output` 包含错误描述。

### Requirement: Glob 工具权限检查钩子

- `Tool` MUST 支持可选的 `PermissionProvider` 注入（通过 `New(perms)` 构造函数）。
- 当 `PermissionProvider` 非 nil 时，`Execute` MUST 在执行搜索前调用 `Check(ctx, "glob:execute", pattern)`。
- 当 `PermissionProvider.Check` 返回非 nil error 时，`Execute` MUST 不执行搜索，直接返回 `ToolResult{IsError: true, Output: error message}`。
- 当 `PermissionProvider` 为 nil 时，`Execute` MUST 跳过权限检查，直接执行搜索。
