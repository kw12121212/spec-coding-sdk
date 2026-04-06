# registry-tasks

## What

为项目增加任务注册表的首个可交付版本，定义最小可落地的任务生命周期能力。

该变更将补充一个新的 task-registry 规格，覆盖以下可观察行为：
- 调用方提供任务 ID 创建任务
- 按 ID 读取已有任务
- 更新任务标题、元数据和状态
- 删除任务时将任务标记为 `deleted`
- 校验合法状态流转和常见错误场景

## Why

M7 是当前主线 roadmap 中下一个基础能力里程碑，而 `registry-tasks` 是该里程碑中依赖最少、可复用性最高的切入口。

先定义任务注册表的最小闭环有三个直接价值：
- 为后续 `registry-teams` 提供同类注册表的行为模式
- 为 `registry-cron` 提供状态管理和错误语义的参考
- 让 README 中已声明的 task registry 能力开始具备明确的规格边界，而不是停留在占位描述

## Scope

In scope:
- 新增任务注册表规格文件，定义创建、读取、更新、删除和状态流转行为
- 限定 v1 为最小可落地范围，不包含枚举任务集合的能力
- 明确任务字段边界：`id`、`title`、`status`、`created_at`、`updated_at`，以及可选 `metadata`
- 明确错误语义：空 ID、重复创建、未知任务、非法状态流转

Out of scope:
- 团队注册表
- Cron 注册表
- 持久化存储
- 自动生成任务 ID
- 过滤、分页或排序能力
- 超出最小状态机的扩展状态

## Unchanged Behavior

Behaviors that must not change as a result of this change (leave blank if nothing is at risk):
- 现有 M1-M6 已完成规格保持不变
- 现有工具能力、agent 生命周期、权限模型和 LLM 后端行为不在本变更内
- 本变更不会为 SDK、JSON-RPC、HTTP 或 gRPC 接口层新增任何公开传输契约
