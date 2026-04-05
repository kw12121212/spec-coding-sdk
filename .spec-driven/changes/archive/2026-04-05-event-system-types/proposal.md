# event-system-types

## What

定义结构化事件系统的核心类型，包含具体事件种类（ToolCallEvent、AgentStateEvent、ErrorEvent 等）、事件发射器接口（EventEmitter）以及事件订阅机制（EventSubscriber），使事件可被实例化、序列化和跨包传递。

## Why

M1 的 Done Criteria 要求"事件类型可被实例化并序列化"。当前 `core-interfaces` 仅定义了通用的 `Event` 结构体（type + payload + timestamp），缺乏具体事件种类和发射/订阅机制。后续 M3（Agent 生命周期）、M4（LLM 流式响应）等里程碑都需要依赖这些事件类型来报告状态变化和工具调用结果。完成此 change 即可关闭 M1。

## Scope

- 在 `internal/core/` 中定义具体事件类型：`ToolCallEvent`、`ToolResultEvent`、`AgentStateEvent`、`ErrorEvent`
- 定义 `EventEmitter` 接口（发射事件）
- 定义 `EventSubscriber` 接口（订阅事件，通过回调函数）
- 确保所有事件类型可序列化为 JSON 并可从 JSON 反序列化
- 所有类型可被其他包引用
- 更新 `project-structure.md` spec 补充事件系统需求

## Unchanged Behavior

- 已有的 `Event` 结构体字段和 JSON tag 不变
- 已有的 `Tool`、`Agent`、`PermissionProvider` 接口不变
- 已有的 `Config` 和 `LoadConfig` 不变
- `pkg/sdk/` 和 `api/proto/` 不被修改
