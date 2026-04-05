# Design: llm-streaming

## Approach

### 1. 通用 SSE 解析器

在 `internal/llm/streaming/` 中实现 `SSEParser`，从 `io.Reader` 读取 SSE 流，产出结构化事件：

```
type SSEEvent struct {
    Event string // event: 行的值（无 event: 行时为空）
    Data  string // data: 行的内容
}

type SSEParser struct { ... }

func NewSSEParser(r io.Reader) *SSEParser
func (p *SSEParser) Next() (SSEEvent, error) // 返回下一个事件；io.EOF 表示流结束
```

解析规则：
- `event: <type>` 行设置当前事件类型
- `data: <payload>` 行触发一个事件产出
- 空行重置事件类型
- 支持 OpenAI 的 `data: [DONE]` 终止信号（返回 io.EOF）
- 支持 Claude 的 `event: message_stop` 终止信号（返回 io.EOF）

### 2. 流式 Tool Call 累积器

在 `internal/llm/streaming/` 中实现 `ToolCallAccumulator`，处理跨 chunk 的增量 tool call 数据拼接：

```
type ToolCallAccumulator struct { ... }

func NewToolCallAccumulator() *ToolCallAccumulator
func (a *ToolCallAccumulator) Feed(chunk StreamChunk) []ToolCall
```

- 对于 OpenAI：每个 delta chunk 的 `tool_calls[i].function.arguments` 是部分 JSON 字符串，累积器按 index 缓存并拼接
- 对于 Claude：`input_json_delta` 的 `partial_json` 是部分 JSON 字符串，累积器按 content block index 缓存并拼接
- `Feed()` 返回已经完成的（全部拼完的）`ToolCall` 列表，未完成的不返回
- 流结束时，`Flush()` 返回所有累积的 tool call（即使不完整也返回，供错误处理）

### 3. 重构 Provider Stream 方法

- OpenAI `Stream()`: 使用 `SSEParser` 替换手动 `bufio.Scanner` 行解析，使用 `ToolCallAccumulator` 拼接 tool call
- Claude `Stream()`: 使用 `SSEParser` 替换手动行解析，使用 `ToolCallAccumulator` 拼接 tool call

两种 provider 在重构后行为不变，但内部实现复用公共组件。

## Key Decisions

1. **SSEParser 放在 `internal/llm/streaming/` 而非 `internal/llm/`**：SSE 解析是流式处理的具体实现细节，不属于 provider-agnostic 接口层。`internal/llm/` 只保留接口和类型定义。

2. **累积器只返回完整 tool call**：`Feed()` 在累积完成前不返回中间状态，调用方无需处理 partial JSON。这简化了 consumer 侧逻辑，但意味着调用方在流中间不会看到不完整的 tool call。

3. **保持 StreamChunk 回调模式不变**：不引入 channel 或 async 模式，保持与现有 `StreamCallback` 签名一致。

## Alternatives Considered

1. **在每个 provider 内部各自实现累积**：被否决，因为两种 provider 的累积逻辑高度相似（都是按 index 缓存 partial JSON），提取为公共组件可避免重复并保证行为一致。

2. **使用 channel 替代 callback**：被否决，channel 模式增加 goroutine 管理复杂度，且与现有 `llm.StreamCallback` 签名不兼容。

3. **将 SSE 解析器作为独立通用库（如 `internal/sse/`）**：被否决，当前项目只有 LLM 调用需要 SSE 解析，放在 `internal/llm/streaming/` 下更内聚。若未来有其他 SSE 场景再提取。
