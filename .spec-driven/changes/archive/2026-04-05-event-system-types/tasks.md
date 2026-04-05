# Tasks: event-system-types

## Implementation

- [x] 在 `internal/core/` 新增 `events.go`，定义事件类型常量（`EventToolCall`、`EventToolResult`、`EventAgentState`、`EventError`）
- [x] 定义具体事件结构体：`ToolCallEvent`、`ToolResultEvent`、`AgentStateEvent`、`ErrorEvent`，每个包含对应业务字段和 JSON tag
- [x] 为每个具体事件类型实现 `EventType() string` 方法，返回对应常量
- [x] 定义 `EventEmitter` 接口，包含 `Emit(event Event)` 方法
- [x] 定义 `EventSubscriber` 接口，包含 `Subscribe(eventType string, handler func(Event))` 方法
- [x] 更新 `.spec-driven/specs/project-structure.md`，补充事件系统类型的需求描述

## Testing

- [x] `make lint` 通过
- [x] `make test` 通过
- [x] 测试各具体事件类型可实例化并字段正确
- [x] 测试各具体事件类型 JSON 序列化/反序列化 round-trip
- [x] 测试外部包可实现 `EventEmitter` 接口
- [x] 测试外部包可实现 `EventSubscriber` 接口
- [x] `go build ./...` 通过

## Verification

- [x] 验证实现与 proposal 范围一致
- [x] 验证已有接口和类型未被修改
- [x] 验证 `pkg/sdk/` 和 `api/proto/` 未被修改
