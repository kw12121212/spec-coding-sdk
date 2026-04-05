# Design: llm-provider-interface

## Approach

在 `internal/llm/` 包中定义 LLM 调用层的所有抽象类型和接口。包结构为 `internal/llm/provider.go`（接口和核心类型）。类型设计参考 OpenAI Chat Completion API 和 Anthropic Messages API 的共同子集，确保两种格式都能无信息损失地映射到统一类型。

核心思路：
- `LLMMessage` 覆盖 user、assistant、tool 三种角色，与 agent 包的 `Role` 对齐
- `LLMToolCall` 字段与 orchestrator 的 `ToolCall` 一致（name + json.RawMessage input）
- `Provider` 接口提供 `Complete`（同步）和 `Stream`（流式）两个方法
- `Stream` 通过回调函数逐 chunk 推送，避免暴露 channel 细节

## Key Decisions

1. **独立 `internal/llm/` 包** — LLM 调用是独立关注点，不放在 `internal/core/`（太泛）也不放在 `internal/agent/`（太窄）。agent 包的 `Thinker` 接口通过适配器模式桥接到 `Provider`。

2. **`Provider` 接口不包含配置** — 配置通过构造函数注入到具体 provider 实现，接口只关心调用行为。这样不同 provider 可以有各自独特的配置参数。

3. **`Stream` 用回调而非 channel** — 回调模式更简单，provider 实现内部自行管理 SSE 解析和 goroutine。channel 模式需要调用方处理关闭和错误，增加复杂度。

4. **`LLMMessage` 包含可选的 `ToolCalls` 字段** — 仅 assistant 角色使用，其他角色为零值。避免定义多种消息子类型。

5. **`LLMUsage` 包含 PromptTokens 和 CompletionTokens** — 两种 API 都提供这两个维度的用量数据。不做更细粒度的拆分（如 Anthropic 的 cache read/write），这些可以在具体 provider 中扩展。

## Alternatives Considered

1. **将 LLM 类型放在 `internal/core/`** — 被否决。core 包应保持最小通用类型，LLM 是特定领域。

2. **定义 `Provider` 接口包含配置和生命周期（Init/Close）** — 被否决。配置和初始化属于构造阶段，不需要在接口中体现。Close 行为由具体实现决定（如连接池、HTTP client 的生命周期）。

3. **流式用 `<-chan Chunk` 而非回调** — 被否决。channel 需要 provider 端管理 goroutine 和 channel 关闭，调用方需要 select + error 处理。回调模式语义更直接，与 SSE 逐行解析天然匹配。
