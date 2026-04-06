# Questions: tool-gh

## Open

<!-- No open questions -->

## Resolved

- [x] Q: `tool-gh` 是否应暴露任意 `gh` 参数，还是只支持受限的 GitHub 操作集合？
  Context: 这决定输入 schema 是通用 CLI 包装还是高层命令抽象。
  A: 暴露任意 `gh` 参数，使用结构化 `args []string` 输入。（用户确认接受建议）

- [x] Q: `tool-gh` 是否需要支持单次调用级别的环境变量覆盖？
  Context: 这影响认证与配置的输入模型，以及权限边界。
  A: 不支持单次调用 env overrides，仅继承当前进程环境。（用户确认接受建议）
