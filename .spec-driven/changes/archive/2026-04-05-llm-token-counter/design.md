# Design: llm-token-counter

## Approach

采用字符启发式估算 + 静态模型注册表方案，不引入外部 tokenizer 依赖。

1. **Token 估算** — 使用 `EstimateTokens(messages []Message) int` 函数，按 ~4 字符/token 的启发式比例估算文本 token 数。Tool call 的 JSON 序列化结果也计入估算。这个比例对英文和代码接近准确，对中文偏低但足够用于窗口检查。

2. **模型注册表** — 定义 `ModelContextWindow` 结构体（含 `TotalTokens` 和 `ModelID`），通过包级 map 提供按模型 ID 查询上下文窗口大小的能力。覆盖主流 OpenAI 和 Claude 模型的已知窗口大小。

3. **上下文检查** — 定义 `ContextChecker` 结构体，组合 token 估算和模型注册表，提供 `Fits(messages []Message, model string, reserved int) bool` 和 `Remaining(messages []Message, model string, reserved int) (int, error)` 方法。`reserved` 参数为响应预留的 token 数。

## Key Decisions

- **启发式估算而非 tiktoken** — tiktoken-go 依赖 CGO，增加构建复杂度。对于上下文窗口检查场景，~15% 的估算误差完全可接受（窗口大小通常有数量级差异）。**Why:** 保持纯 Go 构建，避免交叉编译问题。
- **包级 map 注册表** — 使用包级变量存储模型窗口大小，提供查询函数而非可变注册 API。**Why:** 模型规格是已知常量，不需要运行时动态注册；未来新增模型只需在 map 中添加条目。
- **放置在 `internal/llm/` 包内** — token 估算和上下文检查直接操作 `Message` 类型，放在同一包内避免循环依赖。**Why:** 保持 `internal/llm/` 作为 LLM 调用层的完整抽象边界。

## Alternatives Considered

- **tiktoken-go / customized tokenizers** — 精确计数但引入 CGO 依赖，增加构建和分发成本。被否决因为精确性对窗口检查非必需。
- **API 返回的 Usage 作为唯一 token 来源** — 只能事后获取，无法在发送前判断。被否决因为不满足"发送前估算"需求。
- **独立 `internal/llm/tokenizer/` 包** — 过度拆分，估算逻辑简单，不值得单独包。被否决因为 YAGNI。
