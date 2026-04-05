# Tasks: llm-provider-interface

## Implementation

- [x] 创建 `internal/llm/` 包目录
- [x] 在 `internal/llm/provider.go` 中定义 `LLMMessage` 结构体（Role、Content、ToolName、ToolCalls 字段）
- [x] 在 `internal/llm/provider.go` 中定义 `LLMRole` 类型和常量（LLMRoleUser、LLMRoleAssistant、LLMRoleTool）
- [x] 在 `internal/llm/provider.go` 中定义 `LLMToolCall` 结构体（ID、Name、Input 字段）
- [x] 在 `internal/llm/provider.go` 中定义 `LLMRequest` 结构体（Model、Messages、Temperature、MaxTokens、Tools 字段）
- [x] 在 `internal/llm/provider.go` 中定义 `LLMResponse` 结构体（Content、ToolCalls、Usage、StopReason 字段）
- [x] 在 `internal/llm/provider.go` 中定义 `LLMUsage` 结构体（PromptTokens、CompletionTokens 字段）
- [x] 在 `internal/llm/provider.go` 中定义 `StreamCallback` 函数类型
- [x] 在 `internal/llm/provider.go` 中定义 `Provider` 接口（Complete + Stream 方法）
- [x] 在 `internal/llm/provider.go` 中定义 `ProviderConfig` 结构体（BaseURL、APIKey、Model 字段）
- [x] 确保所有类型可 JSON 序列化/反序列化
- [x] 确保 `Provider` 接口可由外部包实现

## Testing

- [x] `LLMMessage` JSON round-trip 测试
- [x] `LLMRequest` JSON round-trip 测试
- [x] `LLMResponse` JSON round-trip 测试
- [x] `LLMUsage` JSON round-trip 测试
- [x] 外部包可实现 `Provider` 接口的编译期验证
- [x] Lint passes
- [x] `go build ./...` passes

## Verification

- [x] `go test ./internal/llm/...` passes
- [x] `make lint` passes
- [x] Delta spec 与实际实现一致
