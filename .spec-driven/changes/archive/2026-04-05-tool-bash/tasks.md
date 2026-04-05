# Tasks: tool-bash

## Implementation

- [x] 创建 `internal/tools/bash/` 包，定义 `BashInput` 结构体（Command string, Timeout int, WorkingDir string）
- [x] 实现 `BashTool` 结构体，含 `PermissionProvider` 字段（可选）和 `Execute` 方法
- [x] 在 `Execute` 中实现权限检查钩子（PermissionProvider 非 nil 时调用 Check）
- [x] 通过 `exec.CommandContext` + `/bin/bash -c` 执行命令，合并捕获 stdout+stderr
- [x] 实现超时控制（默认 120s，可通过 BashInput.Timeout 覆盖，使用 context.WithTimeout）
- [x] 实现输出大小限制（1MB，超出截断并设置 IsError=true）
- [x] 处理超时和命令不存在的错误场景，返回合适的 ToolResult

## Testing

- [x] 测试正常命令执行（echo、简单 shell 表达式）
- [x] 测试 stderr 输出被合并到 ToolResult.Output
- [x] 测试超时场景（短超时 + sleep 命令）
- [x] 测试输出超过大小限制的场景
- [x] 测试权限被拒绝的场景（PermissionProvider.Check 返回 error）
- [x] 测试 WorkingDir 设置
- [x] 测试无效输入（空 command、非法 JSON）
- [x] `make lint` 通过
- [x] `make test` 通过

## Verification

- [x] BashTool 满足 `core.Tool` 接口（编译时检查）
- [x] 所有测试独立运行、无共享状态
- [x] `go build ./...` 通过
