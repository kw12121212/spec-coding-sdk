# Design: tool-bash

## Approach

在 `internal/tools/bash/` 新建包，定义 `BashTool` 结构体和 `BashInput` 输入结构体。

执行流程：
1. 将 `json.RawMessage` 反序列化为 `BashInput`
2. 若配置了 `PermissionProvider`，调用 `Check(ctx, "bash:execute", command)`
3. 通过 `exec.CommandContext` 启动 bash 子进程，传入命令
4. 捕获 stdout 和 stderr 到统一 buffer（合并输出，与 claw-code 行为对齐）
5. 等待命令完成或 context 超时
6. 返回 `ToolResult{Output: combined, IsError: exitCode != 0 或超时}`

## Key Decisions

1. **合并 stdout/stderr** — 与 claw-code 行为对齐，agent 通常不区分两个流
2. **默认超时 120 秒** — 与 claw-code 默认值对齐，可通过 input 的 `timeout` 字段覆盖
3. **输出大小限制 1MB** — 防止单次工具调用消耗过多内存，超出时截断并标记 IsError
4. **通过 `/bin/bash -c` 执行** — 支持管道、重定向等 shell 特性，与 claw-code 对齐
5. **权限检查为可选** — `PermissionProvider` 通过选项模式注入，nil 时跳过检查（测试友好）
6. **working_dir 可选** — 默认继承进程工作目录，可在 input 中指定

## Alternatives Considered

1. **分离 stdout/stderr** — 增加了接口复杂度但 agent 很少需要区分，排除
2. **直接 exec 命令（不走 bash -c）** — 无法支持管道和重定向，排除
3. **流式输出返回** — 当前 `ToolResult` 是一次性返回，流式需要修改核心接口或引入回调，留到后续迭代
