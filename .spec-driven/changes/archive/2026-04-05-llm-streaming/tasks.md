# Tasks: llm-streaming

## Implementation

- [x] 实现 `internal/llm/streaming/sse.go`：`SSEEvent` 类型、`SSEParser` 结构体、`NewSSEParser()`、`Next()` 方法
- [x] 实现 `internal/llm/streaming/accumulator.go`：`ToolCallAccumulator` 结构体、`NewToolCallAccumulator()`、`Feed()` 方法、`Flush()` 方法
- [x] 重构 `internal/llm/openai/provider.go` 的 `Stream()` 方法，使用 `SSEParser` 和 `ToolCallAccumulator`
- [x] 重构 `internal/llm/claude/provider.go` 的 `Stream()` 方法，使用 `SSEParser` 和 `ToolCallAccumulator`
- [x] 更新 `llm-backend.md` delta spec，补充流式处理相关需求

## Testing

- [x] `internal/llm/streaming/sse_test.go`：SSE 解析器单元测试（正常事件流、带 event 类型的流、`[DONE]` 终止、空行分隔、格式错误处理）
- [x] `internal/llm/streaming/accumulator_test.go`：累积器单元测试（单 chunk 完整 tool call、多 chunk 拼接、多个并发 tool call、flush 未完成数据）
- [x] `internal/llm/openai/provider_test.go`：确保重构后现有测试全部通过，补充流式 tool call 跨 chunk 测试
- [x] `internal/llm/claude/provider_test.go`：确保重构后现有测试全部通过，补充流式 tool call 跨 chunk 测试
- [x] 全项目 `go vet` 和 `go build` 通过

## Verification

- [x] 所有新增和既有测试通过
- [x] `go vet ./...` 无警告
- [x] delta spec 内容与实际实现一致
