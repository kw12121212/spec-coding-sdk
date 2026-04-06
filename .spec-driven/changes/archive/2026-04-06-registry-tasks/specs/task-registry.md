# task-registry

## ADDED Requirements

### Requirement: Empty task registry initialization

- 项目 MUST 提供一个新的任务注册表实例，且该实例在创建后初始不包含任何任务。
- 当调用方在创建任何任务前按 ID 读取任务时，系统 MUST 返回非 nil error。

### Requirement: Task creation with caller-supplied ID

- 调用方 MUST 能够使用非空任务 ID 创建任务。
- 当调用方创建任务时，系统 MUST 要求任务标题为非空字符串。
- 当调用方成功创建任务时，系统 MUST 保存以下字段：`id`、`title`、`status`、`created_at`、`updated_at`。
- 新建任务的 `status` MUST 为 `pending`。
- 当调用方在创建时提供 `metadata` 时，系统 MUST 原样保存该值。
- 当调用方使用空 ID 或空标题创建任务时，系统 MUST 返回非 nil error。
- 当调用方使用已存在的任务 ID 再次创建任务时，系统 MUST 返回非 nil error，且 MUST NOT 覆盖既有任务。

### Requirement: Task retrieval and mutation by ID

- 调用方 MUST 能够按任务 ID 读取已存在的任务。
- 调用方 MUST 能够按任务 ID 更新已存在任务的 `title`、`metadata` 和 `status`。
- 调用方 MUST 能够按任务 ID 删除已存在任务。
- 当调用方读取、更新或删除不存在的任务 ID 时，系统 MUST 返回非 nil error。
- 当任务更新成功时，系统 MUST 保留原有 `created_at`，并更新 `updated_at`。
- 当任务更新标题或元数据时，系统 MUST 在后续读取中返回更新后的值。
- 当调用方删除任务时，系统 MUST 将该任务的 `status` 设为 `deleted`，而不是将其从注册表中移除。
- 当删除成功后再次按相同 ID 读取任务时，系统 MUST 返回该任务，且其 `status` 为 `deleted`。

### Requirement: Task status transition validation

- 系统 MUST 仅允许以下状态流转：`pending -> in_progress`、`pending -> deleted`、`in_progress -> completed`、`in_progress -> deleted`。
- 当调用方请求未列出的状态流转时，系统 MUST 返回非 nil error。
- 当状态流转失败时，系统 MUST NOT 改变任务原有状态。
- 处于 `completed` 状态的任务 MUST NOT 再次变更为其他状态。
- 处于 `deleted` 状态的任务 MUST NOT 再次变更为其他状态。
