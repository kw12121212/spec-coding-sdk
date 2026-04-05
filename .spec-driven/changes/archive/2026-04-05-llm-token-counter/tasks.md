# Tasks: llm-token-counter

## Implementation

- [x] 实现 `EstimateTokens(messages []Message) int` 函数（字符启发式 ~4 chars/token，含 tool call JSON 估算）
- [x] 定义 `ModelContextWindow` 结构体和包级模型注册表 map（覆盖 gpt-4o、gpt-4-turbo、gpt-3.5-turbo、claude-sonnet-4-6、claude-opus-4-6、claude-haiku-4-5 等主流模型）
- [x] 实现 `ContextWindow(modelID string) (int, bool)` 查询函数
- [x] 实现 `ContextChecker` 结构体及 `Fits` / `Remaining` 方法
- [x] 更新 `llm-backend.md` delta spec，补充 token 计数相关需求

## Testing

- [x] Lint passes
- [x] `EstimateTokens` 测试：空消息列表、纯文本消息、含 tool call 消息、多消息混合
- [x] `ContextWindow` 测试：已知模型返回正确窗口大小、未知模型返回 false
- [x] `ContextChecker` 测试：fits/remaining 正常、超窗口、reserved 参数、未知模型处理

## Verification

- [x] Verify implementation matches proposal scope
- [x] Verify delta specs reflect actual implementation
