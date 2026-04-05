# Tasks: llm-provider-openai

## Implementation

- [x] 创建 `internal/llm/openai/` 包目录
- [x] 实现 OpenAI 请求格式转换（内部 Message → OpenAI JSON）
- [x] 实现 OpenAI 响应格式转换（OpenAI JSON → 内部 Response / StreamChunk）
- [x] 实现 `OpenAIProvider` struct 和 `NewProvider` 构造函数（含 `WithHTTPClient` option）
- [x] 实现 `Complete` 方法（HTTP POST + JSON 解析）
- [x] 实现 `Stream` 方法（SSE 解析 + chunk 回调）
- [x] 实现错误处理（HTTP 错误、API 错误响应）
- [x] 添加编译期接口满足检查

## Testing

- [x] 测试同步调用正常文本响应
- [x] 测试同步调用工具调用响应
- [x] 测试流式调用正常文本
- [x] 测试 HTTP 错误响应处理
- [x] 测试空消息列表边界情况
- [x] 测试 BaseURL 配置（兼容供应商）

## Verification

- [x] 所有测试通过（`go test ./internal/llm/...`）
- [x] `go vet` 无警告
- [x] 已有测试不受影响
