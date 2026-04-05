# llm-streaming

## What

在 `internal/llm/streaming/` 包中实现统一的 SSE 流式解析器，将 OpenAI 和 Claude 两种 provider 的流式响应解析逻辑提取为可复用的公共组件。同时为两种 provider 的 Stream 方法实现流式工具调用增量累积，确保跨 chunk 的 tool call 数据可被调用方完整消费。

## Why

当前 OpenAI 和 Claude 两个 provider 各自在 `Stream()` 方法中内嵌了 SSE 解析逻辑（`bufio.Scanner` + `data:` 行解析），存在以下问题：
1. SSE 解析代码重复，新增 provider 时需要再次复制
2. Claude 的 `input_json_delta` 流式 tool call 输入是分多个 chunk 逐片发送的，当前实现只传递 `partial_json` 原始片段，调用方需要自行拼接才能获得完整 JSON
3. OpenAI 的流式 tool call 的 `function.arguments` 也可能跨多个 delta chunk 分片到达
4. 缺乏流式场景下的错误恢复机制（网络中断、格式错误等）

## Scope

**In Scope**:
- `internal/llm/streaming/` 包：通用 SSE 行解析器
- `internal/llm/streaming/` 包：流式 tool call 增量累积器，自动拼接跨 chunk 的 partial JSON
- 重构 OpenAI provider 的 `Stream()` 方法，使用公共 SSE 解析器
- 重构 Claude provider 的 `Stream()` 方法，使用公共 SSE 解析器
- 两者的 Stream 方法通过累积器交付完整的 `ToolCall.Input` JSON

**Out of Scope**:
- 上下文压缩/截断策略
- 非 OpenAI/Claude 格式的 provider
- 请求重试逻辑（属于独立关注点）
- Token 计数（`llm-token-counter` 变更）

## Unchanged Behavior

- `llm.Provider` 接口签名不变
- `llm.StreamChunk` 类型定义不变
- `OpenAIProvider.Complete()` 和 `ClaudeProvider.Complete()` 行为不变
- 已有的 `Stream()` 方法对外可观察行为不变（相同的 chunk 序列、相同的回调语义）
