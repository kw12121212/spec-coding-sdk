# Questions: tool-bash

## Open

<!-- No open questions -->

## Resolved

- [x] Q: Bash 工具是否需要支持交互式命令（如 `gcloud auth login`）？
  Context: 交互式命令需要 pty 支持，影响实现复杂度和接口设计
  A: 不支持。仅支持非交互式执行。（用户确认）

- [x] Q: 超时默认值是多少？
  Context: 影响工具行为和测试设计
  A: 默认 120 秒，与 claw-code 对齐，可通过 input.Timeout 覆盖。超时强制杀死子进程。对于服务器类挂起命令，agent 可通过调高 timeout 值应对，后台执行模式不在当前 scope。（用户确认）
