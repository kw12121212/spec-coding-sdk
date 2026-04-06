# Questions: registry-tasks

## Open

<!-- No open questions -->

## Resolved

- [x] Q: v1 是否包含 `List` 能力？
  Context: 这会决定任务注册表是只覆盖按 ID 的最小闭环，还是在首版就引入集合语义。
  A: 不包含，采用最小可落地版，只做按 ID 的创建、读取、更新、删除。

- [x] Q: 任务 ID 由调用方提供还是由系统自动生成？
  Context: 这会决定注册表是否承担资源创建策略，以及重复创建的边界如何定义。
  A: 任务 ID 由调用方提供，缺失 ID 视为错误。

- [x] Q: v1 的最小任务字段是什么？
  Context: 这会决定规格中哪些字段必须稳定可观察，哪些能力应留待后续扩展。
  A: 最小字段为 `id`、`title`、`status`、`created_at`、`updated_at`，`metadata` 为可选。

- [x] Q: 哪些状态流转是合法的？
  Context: 这会决定状态机边界、错误语义以及后续注册表扩展的兼容性。
  A: 合法流转为 `pending -> in_progress`、`pending -> deleted`、`in_progress -> completed`、`in_progress -> deleted`；`completed` 和 `deleted` 为终态。
