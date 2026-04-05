# M2 - Tool Surface 基础工具集

## Goal

实现 bash、文件操作、grep、glob 四大核心工具，构成 agent 可调用的基础 tool surface。

## In Scope

- Bash 执行工具（命令运行、超时、输出捕获）
- 文件读/写/编辑工具
- 内容搜索工具（基于 ripgrep）
- 文件模式匹配工具（glob pattern）
- 工具级别预留权限检查钩子（基于 M1 PermissionProvider 接口）

## Out of Scope

- 内置外部工具管理（M2x）
- LSP 客户端工具（M8）
- MCP 协议工具（M9）
- 权限策略的具体执行逻辑（M5）

## Done Criteria

- 每个工具可独立实例化并执行基本操作
- 每个工具有对应的单元测试覆盖 happy path 和 error case
- 工具输入/输出符合 M1 定义的 Tool 接口
- 每个工具在执行前可通过 PermissionProvider 钩子进行权限检查

## Planned Changes

- `tool-bash` - Bash 命令执行工具实现
- `tool-file-ops` - 文件读/写/编辑工具实现
- `tool-grep` - 基于 ripgrep 的内容搜索工具实现
- `tool-glob` - 文件模式匹配工具实现

## Dependencies

- M1 核心接口（Tool 接口、PermissionProvider 接口）

## Risks

- Bash 工具的安全性约束需要仔细设计，避免命令注入
- 文件编辑操作的原子性和并发安全

## Status

- Declared: proposed

## Notes

- 工具行为应与 claw-code Rust 实现保持**功能对等、设计对齐但不照搬**
- tool-grep 使用的 ripgrep 二进制由 M2x 的 builtin-tool-manager 提供
