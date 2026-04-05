# Tasks: tool-file-ops

## Implementation

- [x] 创建 `internal/tools/fileops/` 包，定义 `ReadInput`、`WriteInput`、`EditInput` 输入结构体
- [x] 实现 `ReadTool` 结构体：文件读取、行号格式化、offset/limit 截取
- [x] 实现 `WriteTool` 结构体：父目录自动创建（`os.MkdirAll`）、文件写入
- [x] 实现 `EditTool` 结构体：精确字符串查找、唯一性验证、替换执行
- [x] 三个工具均实现 `PermissionProvider` 可选注入和执行前权限检查
- [x] 绝对路径验证逻辑（共享辅助函数）

## Testing

- [x] ReadTool：正常读取、行号格式验证、offset 截取、limit 截取、文件不存在、路径为目录
- [x] ReadTool：空 file_path、非绝对路径、非法 JSON 输入
- [x] WriteTool：正常写入、父目录自动创建、覆盖已存在文件、写入空内容
- [x] WriteTool：空 file_path、非绝对路径、非法 JSON 输入
- [x] EditTool：单次替换成功、多处匹配报错、未找到匹配报错、replace_all 全局替换
- [x] EditTool：空 file_path、空 old_string、非绝对路径、非法 JSON 输入、文件不存在
- [x] 三个工具各自的权限拒绝和权限 nil 跳过测试
- [x] `make lint` 通过
- [x] `make test` 通过

## Verification

- [x] ReadTool、WriteTool、EditTool 满足 `core.Tool` 接口（编译时检查）
- [x] 所有测试独立运行、无共享状态
- [x] `go build ./...` 通过
