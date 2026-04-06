# Design: tool-gh

## Approach

在 `internal/tools/gh/` 中增加一个基于 `gh` CLI 的工具实现，并沿用现有
工具模式：输入为 `json.RawMessage`，输出为 `core.ToolResult`。执行流程：

1. 将输入反序列化为 `Input{Args, WorkingDir, Timeout}`
2. 验证 `args` 非空且不包含空字符串
3. 若配置了 `PermissionProvider`，先对本次命令做权限检查
4. 通过已完成的 built-in tool manager 解析 `gh` 可执行文件路径
5. 使用解析出的可执行文件直接启动子进程，不经过 shell 拼接
6. 继承当前进程环境，按需设置工作目录与超时
7. 合并捕获 stdout/stderr，应用 1MB 输出上限
8. 根据退出码、超时或解析失败结果返回 `ToolResult`

## Key Decisions

1. **输入采用 `args []string` 而非单个命令字符串**
   Rationale: `gh` 本身是结构化 CLI，参数数组比 shell 字符串更稳定，也避免
   在 spec 中引入额外的 shell 解析行为。

2. **每次执行都通过 built-in manager 解析 `gh`**
   Rationale: 这样调用方无需区分系统安装与托管安装，工具行为与 M3 已完成的
   能力保持一致。

3. **继承当前进程环境，不提供单次调用 env overrides**
   Rationale: 这满足最小可用场景，也避免把认证、配置和环境拼装逻辑扩展到当前
   proposal 范围之外。

4. **不支持交互式命令**
   Rationale: 当前 `Tool` 接口是一次性请求/响应模型，没有 pty 或流式交互能力。
   对 `gh auth login` 之类流程的支持应单独设计。

5. **沿用现有命令工具的超时与输出限制约定**
   Rationale: 这能让 `tool-gh` 的行为和 `tool-bash` 保持一致，降低后续 SDK
   与接口层的差异化处理成本。

## Alternatives Considered

1. **将常见 GitHub 操作建模为固定字段**
   例如把 issue、PR、repo 等命令拆成专用 schema。这样会在没有明确需求前
   引入较重的 API 设计负担，因此排除。

2. **直接依赖 PATH 查找 `gh`，不接 built-in manager**
   这样会绕过 M3 刚完成的基础设施，使“宿主缺失但托管安装可用”的行为无法
   达成，因此排除。

3. **允许单次调用传入环境变量覆盖**
   这会扩大输入 schema 和权限/安全讨论范围，且当前没有明确 roadmap 要求，
   因此排除。
