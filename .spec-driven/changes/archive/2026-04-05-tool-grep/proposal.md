# tool-grep

## What

实现基于 ripgrep 的内容搜索工具，作为 M2 Tool Surface 基础工具集的一部分。工具封装 `rg` 命令行调用，提供正则表达式搜索、文件类型过滤、输出模式选择等能力。

## Why

内容搜索是 coding agent 的核心能力之一。M2 里程碑要求完成 bash、文件操作、grep、glob 四大基础工具，其中 bash 和文件操作已完成。tool-grep 是完成 M2 的必要步骤。

## Scope

- 在 `internal/tools/grep/` 包中实现 `Tool` 结构体，符合 `core.Tool` 接口
- 支持 ripgrep 常用搜索参数：pattern、path、文件类型过滤、输出模式、上下文行数
- 通过 `PermissionProvider` 钩子进行权限检查
- 单元测试覆盖 happy path、输入校验、权限拒绝和错误场景

## Unchanged Behavior

- 不修改 `internal/core/` 中的任何接口或类型定义
- 不影响已有的 bash、fileops 工具实现
- ripgrep 二进制管理由 M3 builtin-tool-manager 负责，本变更假设 `rg` 在 PATH 中可用
