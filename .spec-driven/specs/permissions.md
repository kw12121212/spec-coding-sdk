# permissions

## Requirements

### Requirement: Decision 类型定义

- 项目 MUST 在 `internal/permission/` 包中定义 `Decision` 类型（基于 string）。
- 项目 MUST 定义以下决策常量：`DecisionAllow`（"allow"）、`DecisionDeny`（"deny"）、`DecisionNeedConfirmation`（"need_confirmation"）。
- `Decision` MUST 实现 `String() string` 方法返回决策的字符串表示。

### Requirement: Rule 结构体

- 项目 MUST 在 `internal/permission/` 包中定义 `Rule` 结构体，包含以下字段：
  - `OperationPattern` (string) — 操作模式，支持 `filepath.Match` 语法（如 `"bash:*"`、`"file:read"`）
  - `ResourcePattern` (string) — 资源模式，空字符串表示匹配所有资源
  - `Decision` (Decision) — 匹配时的决策结果
- `Rule` MUST 提供 `Match(operation, resource string) bool` 方法，分别匹配 operation 和 resource 模式。
- 当 `ResourcePattern` 为空字符串时，MUST 匹配所有资源。
- OperationPattern 匹配 MUST 使用 `filepath.Match` 语义（支持 `*`、`?` 通配符）。
- ResourcePattern 匹配 MUST 使用子串包含语义（`strings.Contains`），即 resource 中包含 ResourcePattern 即为匹配。

### Requirement: Policy 接口

- 项目 MUST 在 `internal/permission/` 包中定义 `Policy` 接口，包含方法 `Evaluate(ctx context.Context, operation string, resource string) Decision`。
- `Policy` 接口 MUST 可由外部包实现（不依赖 `internal/permission/` 的非导出类型）。

### Requirement: StaticPolicy 实现

- 项目 MUST 在 `internal/permission/` 包中提供 `StaticPolicy` 结构体，实现 `Policy` 接口。
- `StaticPolicy` MUST 持有有序 `Rule` 切片和 `DefaultDecision` (Decision) 字段。
- `Evaluate` MUST 按顺序遍历规则，返回第一个匹配的 `Rule.Decision`。
- 当无规则匹配时，MUST 返回 `DefaultDecision`。
- `StaticPolicy` MUST 通过 `NewStaticPolicy(defaultDecision Decision, rules ...Rule) *StaticPolicy` 构造函数创建。

### Requirement: 默认策略

- 项目 MUST 在 `internal/permission/` 包中提供 `DefaultPolicy() *StaticPolicy` 函数，返回预配置的策略。
- `DefaultPolicy()` MUST 返回一个无规则的 `StaticPolicy`，`DefaultDecision` 为 `DecisionAllow`，即允许所有操作（full access）。
- 项目 SHOULD 在 `internal/permission/` 包中提供 `SafePolicy() Policy` 函数，返回预配置的安全策略，包含以下规则（按优先级排列）：
  1. 拒绝匹配 `rm -rf`、`rm -r /*`、`mkfs`、`dd if=`、`:(){ :|:& };:` 模式的 bash 命令（硬编码模式列表）
  2. 拒绝当前工作目录之外的文件写入和编辑操作
  3. 允许所有文件读取操作
  4. 允许所有 bash 执行操作（未被规则 1 拒绝的）
  5. 允许 grep 和 glob 操作
- `SafePolicy()` MUST 在创建时捕获当前进程工作目录作为写操作的安全边界。

### Requirement: PolicyProvider 适配器

- 项目 MUST 在 `internal/permission/` 包中提供 `PolicyProvider` 结构体，实现 `core.PermissionProvider` 接口。
- `PolicyProvider` MUST 通过 `NewPolicyProvider(policy Policy) *PolicyProvider` 构造函数创建。
- `Check(ctx, operation, resource)` MUST 调用 `policy.Evaluate(ctx, operation, resource)` 并映射决策：
  - `DecisionAllow` → 返回 nil
  - `DecisionDeny` → 返回错误，错误消息格式为 `"permission denied: <operation> on <resource>"`
  - `DecisionNeedConfirmation` → 返回错误，错误消息格式为 `"confirmation required: <operation> on <resource>"`
- 项目 MUST 包含编译期接口满足检查 `var _ core.PermissionProvider = (*PolicyProvider)(nil)`。
