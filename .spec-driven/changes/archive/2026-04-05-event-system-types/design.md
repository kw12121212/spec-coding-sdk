# Design: event-system-types

## Approach

在 `internal/core/` 包中新增 `events.go` 文件，定义具体事件类型和发射/订阅接口：

1. **具体事件类型** — 每种事件为独立的结构体，内嵌 `Event` 或持有其字段，确保与已有 `Event` 类型兼容。每种事件有明确的类型标识常量。
2. **EventEmitter 接口** — 提供 `Emit(event Event)` 方法，供 Agent 和 Tool 调用以发射事件。
3. **EventSubscriber 接口** — 提供 `Subscribe(eventType string, handler func(Event))` 方法，支持按事件类型注册回调。
4. **JSON 序列化** — 所有事件类型使用标准 `encoding/json` tag，确保可序列化和反序列化。

## Key Decisions

- **事件类型使用常量字符串** — 如 `EventToolCall = "tool.call"`，避免魔法字符串，便于订阅方按类型过滤。
- **不引入 channel** — 订阅机制使用回调函数而非 Go channel。channel 适合内部管道，回调更灵活且与 EventEmitter/Subscriber 接口解耦。后续里程碑如需 channel 可自行包装。
- **EventEmitter 和 EventSubscriber 为接口** — 而非具体实现。M1 只定义类型，具体实现留到 M3 Agent 生命周期。
- **具体事件类型在 core 包定义** — 而非子包。当前事件种类有限，不值得单独子包。如果后续事件类型膨胀，可再拆分。

## Alternatives Considered

- **使用泛型事件 `Event[T]`** — 过度设计，当前 Go 泛型在序列化场景下增加复杂度，收益不大。
- **事件总线（EventBus）具体实现** — 属于运行时逻辑，超出 M1 范围，留给后续里程碑。
- **将事件类型放在 `pkg/sdk/`** — 事件类型是内部实现细节，应放在 `internal/core/`，后续由 SDK facade 重新导出。
