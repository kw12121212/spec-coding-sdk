# Design: tool-grep

## Approach

采用与 bash/fileops 工具一致的模式：定义 Input 结构体，实现 `core.Tool` 接口的 `Execute` 方法，内部通过 `exec.Command` 调用 `rg` 命令。

工具将 pattern 和 path 映射为 `rg` 的位置参数，其余选项映射为命令行 flag。输出通过 stdout 捕获，stderr 合并到输出中。

## Key Decisions

1. **调用 rg 而非纯 Go 实现** — ripgrep 是行业标准的代码搜索工具，性能和正确性远超纯 Go 正则引擎。M3 将负责 rg 二进制的自动管理。

2. **输出模式参数** — 支持 `content`（默认，显示匹配行）、`files_with_matches`（仅文件路径）、`count`（匹配计数）三种模式，与 claw-code 的 grep 工具对齐。

3. **默认行为** — 默认在当前工作目录搜索，输出匹配行内容及行号，不区分大小写关闭，支持正则。

4. **输出截断** — 复用 bash 工具的 1MB 截断策略，保持工具间行为一致。

5. **权限操作标识** — 使用 `"grep:execute"` 作为权限检查的操作名，与 bash 的 `"bash:execute"` 模式一致。

## Alternatives Considered

- **纯 Go regexp 实现** — 需要自行实现文件遍历、.gitignore 支持等，复杂度高且性能差。排除。
- **Go ripgrep 绑定** — ripgrep 无稳定 C API，grepping 库生态不成熟。排除。
- **grep 命令替代** — POSIX grep 不支持 .gitignore、性能差。排除。
