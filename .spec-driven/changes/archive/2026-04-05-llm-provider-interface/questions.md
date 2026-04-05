# Questions: llm-provider-interface

## Open

<!-- No open questions -->

## Resolved

- [x] Q: `Provider.Stream` 方法签名是 `Stream(ctx context.Context, req LLMRequest, callback StreamCallback) error` 还是返回 `io.ReadCloser` 让调用方自行消费？
  Context: 回调模式更简单但灵活性较低；ReadCloser 模式让调用方控制消费节奏。考虑到后续 `llm-streaming` 变更需要统一处理 SSE，回调模式足够且更易测试。
  A: 使用 callback 模式。
