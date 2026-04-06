# Tasks: tool-gh

## Implementation

- [x] 在 `internal/tools/gh/` 中创建 `Tool` 和 `Input`，定义 `args`、
      `working_dir`、`timeout` 输入结构
- [x] 将 `tool-gh` 的可观察行为写入 change delta spec
- [x] 实现 `gh` 可执行文件解析，优先宿主安装并在缺失时回退到 SDK 托管安装
- [x] 实现 `gh` 子进程执行、工作目录设置、超时控制与 stdout/stderr 合并输出
- [x] 实现 1MB 输出大小限制和超时/启动失败/非零退出码错误处理
- [x] 实现可选权限检查钩子，并在拒绝时短路返回错误结果

## Testing

- [x] 测试系统已安装 `gh` 时的正常执行路径
- [x] 测试 PATH 缺失时通过托管安装执行 `gh` 的回退路径
- [x] 测试权限被拒绝时不会解析或执行命令
- [x] 测试无效输入（空 args、空参数元素、非法 JSON）
- [x] 测试非零退出码、超时和输出截断场景
- [x] `make lint` 通过
- [x] `make test` 通过

## Verification

- [x] `go build ./...` 通过
- [x] `Tool` 满足 `core.Tool` 接口，且测试彼此独立无共享状态
- [x] 验证实现范围仅覆盖 `tool-gh`，不扩展到 `tool-rtk` 或更高层 GitHub 抽象
