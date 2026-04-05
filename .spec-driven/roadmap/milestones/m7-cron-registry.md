# M7 - 定时任务注册表

## Goal

实现 cron 定时任务注册表，支持周期性任务调度和一次性延时任务。

## In Scope

- Cron 表达式解析与调度器
- 定时任务注册表（创建、查询、删除）
- 一次性延时任务（fire-once）
- 任务执行的错误处理与重试

## Out of Scope

- 持久化存储实现（内存调度）
- 分布式调度（单进程内）
- 任务持久化与错过触发的恢复

## Done Criteria

- Cron 任务可按表达式周期性触发
- 一次性任务可在指定时间后触发
- 任务可被注册、查询、取消
- 有单元测试覆盖调度精度和错误场景

## Planned Changes

- `registry-cron` - Cron 调度器与定时任务注册表实现

## Dependencies

- M1 核心接口（基础类型定义）
- M6 注册表（参考注册表实现模式）

## Risks

- Cron 调度精度受 Go runtime timer 精度限制
- 长时间运行进程中的 goroutine 泄漏风险

## Status

- Declared: proposed

## Notes

- 调度器使用 Go 标准库 time.Timer 实现，不引入外部 cron 库
- 参考 claw-code 的 cron 调度行为（最长 7 天自动过期等）
