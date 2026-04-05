# M7 - 任务与团队注册表

## Goal

实现任务注册表和团队注册表，提供结构化的任务追踪和团队协作基础设施。

## In Scope

- 任务注册表（创建、查询、更新、删除、状态流转）
- 团队注册表（成员管理、角色分配、团队创建/解散）
- 内存存储实现（接口预留持久化扩展）

## Out of Scope

- 持久化存储实现
- Cron 定时任务注册表（M8）
- 权限策略（M6）

## Done Criteria

- 任务注册表 CRUD 操作均可正常工作
- 任务状态流转（pending → in_progress → completed/deleted）正确
- 团队注册表的成员加入/移除和角色分配正确
- 有单元测试覆盖边界条件（空 ID、重复创建等）

## Planned Changes

- `registry-tasks` - Declared: planned - 任务注册表实现
- `registry-teams` - Declared: planned - 团队注册表实现

## Dependencies

- M1 核心接口（基础类型定义）

## Risks

- 并发访问时的数据一致性（内存版本可用 mutex 控制复杂度）
- 任务状态机的合法转换路径需要明确定义

## Status

- Declared: proposed

## Notes

- 注册表先实现内存版本，持久化作为后续增强
- 与 M6（权限）无强依赖，可并行开发
