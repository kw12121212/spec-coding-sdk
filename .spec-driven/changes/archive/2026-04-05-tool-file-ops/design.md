# Design: tool-file-ops

## Approach

在 `internal/tools/fileops/` 新建包，定义三个工具结构体：`ReadTool`、`WriteTool`、`EditTool`，共享同一个包但各自独立实现 `core.Tool` 接口。

### ReadTool

执行流程：
1. 反序列化 JSON 输入为 `ReadInput`
2. 验证 `file_path` 为绝对路径
3. 权限检查（`Check(ctx, "file:read", file_path)`）
4. 读取文件内容，按行编号
5. 若指定 `offset`/`limit`，截取对应行范围
6. 返回带行号的内容

### WriteTool

执行流程：
1. 反序列化 JSON 输入为 `WriteInput`
2. 验证 `file_path` 为绝对路径
3. 权限检查（`Check(ctx, "file:write", file_path)`）
4. 自动创建父目录（`os.MkdirAll`）
5. 写入内容到文件（`os.WriteFile`，权限 0644）
6. 返回成功确认

### EditTool

执行流程：
1. 反序列化 JSON 输入为 `EditInput`
2. 验证 `file_path` 为绝对路径
3. 权限检查（`Check(ctx, "file:edit", file_path)`）
4. 读取当前文件内容
5. 查找 `old_string`，验证唯一性（除非 `replace_all=true`）
6. 执行字符串替换
7. 写回文件
8. 返回替换次数

## Key Decisions

1. **三工具分立、同包共存** — 各工具职责清晰、输入 schema 独立，共享 `fileops` 包避免包爆炸
2. **绝对路径强制** — `file_path` 必须为绝对路径，避免 agent 在不确定工作目录时产生歧义
3. **EditTool 精确匹配** — `old_string` 必须与文件中的内容完全匹配（包括空白字符），确保替换的精确性
4. **默认唯一性检查** — `replace_all=false` 时，`old_string` 在文件中必须恰好出现一次，否则返回错误。防止意外多处替换
5. **WriteTool 自动创建父目录** — 与 claw-code 行为对齐，agent 通常期望写入路径可用
6. **权限操作标识** — 分别使用 `file:read`、`file:write`、`file:edit`，粒度与操作类型匹配

## Alternatives Considered

1. **单工具 + operation 字段** — 在一个 Tool 中通过 `operation` 字段区分读/写/编辑。增加了单次输入验证的复杂度，且不同操作的输入 schema 差异较大，排除
2. **行号定位编辑** — 通过行号范围定位编辑区域。行号在文件被并发修改时不可靠，精确字符串匹配更安全，排除
3. **diff/patch 格式编辑** — 使用 unified diff 格式。复杂度高，agent 生成精确 diff 的难度大，排除
4. **分三个包** — `tools/fileread/`、`tools/filewrite/`、`tools/fileedit/`。三个工具紧密相关，共享辅助逻辑，拆包增加维护成本，排除
