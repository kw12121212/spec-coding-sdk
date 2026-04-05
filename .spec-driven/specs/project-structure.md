# project-structure

## ADDED Requirements

### Requirement: Go module

- 项目 MUST 在仓库根目录包含 `go.mod`，模块路径为 `github.com/kw12121212/spec-coding-sdk`，Go 版本不低于 1.25。

### Requirement: 目录布局

- 项目 MUST 包含以下顶层目录：`cmd/`（可执行入口）、`pkg/`（公共库包）、`internal/`（私有实现包）、`api/`（Protobuf 定义及生成代码）。
- 每个顶层目录 MUST 至少包含一个子包，每个子包 MUST 包含至少一个 `.go` 文件，使得 `go build ./...` 无错误通过。初始子包为 `pkg/sdk`、`internal/core`、`api/proto`。

### Requirement: 构建目标

- `make build` MUST 成功编译所有包。
- `make test` MUST 运行所有测试并通过。
- `make lint` MUST 使用 golangci-lint 检查代码，初始配置无错误。
- `make fmt` MUST 使用 `gofmt` 格式化所有 `.go` 文件。
- `make clean` MUST 清除构建产物。

### Requirement: 可执行入口

- `cmd/spec-coding-sdk/main.go` MUST 包含一个 `main` 函数，程序可正常编译运行并退出。

### Requirement: 核心接口定义

- 项目 MUST 在 `internal/core/` 包中定义以下核心接口：
  - `Tool` — 工具调用合约，MUST 包含 `Execute` 方法
  - `Agent` — Agent 生命周期合约，MUST 支持会话管理（启动、停止）和工具调用
  - `PermissionProvider` — 权限检查合约，MUST 包含 `Check` 方法
- 项目 MUST 在 `internal/core/` 包中定义 `Event` 结构体，包含类型标识、负载和时间戳字段
- 项目 MUST 定义 `ToolResult` 类型，作为 `Tool.Execute` 的返回值
- 所有核心类型 MUST 可被其他包引用（`go build ./...` 通过）
- `pkg/sdk/` 和 `api/proto/` 包 MUST NOT 被此变更修改

### Requirement: 接口可实现性

- 每个核心接口 MUST 可由外部包的类型实现（不依赖 `internal/core/` 的非导出类型）
- 测试 MUST 验证其他包可以定义实现这些接口的具体类型

### Requirement: Configuration loading

- 项目 MUST 在 `internal/core/` 包中提供 `LoadConfig(path string) (*Config, error)` 函数，读取指定路径的 YAML 文件并返回解析后的 `Config` 结构体。
- `Config` 结构体 MUST 定义在 `internal/core/` 包中，初始版本不包含具体字段，但 MUST 可被其他包引用。
- 当配置文件不存在或 YAML 格式无效时，`LoadConfig` MUST 返回非 nil 的 error，error 信息 MUST 包含文件路径或具体解析错误。
- `LoadConfig` MUST 在返回 `*Config` 前对解析结果进行验证。初始版本的验证仅确认 YAML 语法正确且文件可读取。

### Requirement: 事件系统类型

- 项目 MUST 在 `internal/core/` 包中定义以下具体事件类型：
  - `ToolCallEvent` — 工具调用事件，MUST 包含工具名称（`ToolName`）和输入参数（`Input`）字段
  - `ToolResultEvent` — 工具结果事件，MUST 包含工具名称（`ToolName`）和执行结果（`Result`）字段
  - `AgentStateEvent` — Agent 状态变化事件，MUST 包含状态标识（`State`）和描述（`Message`）字段
  - `ErrorEvent` — 错误事件，MUST 包含错误码（`Code`）和错误信息（`Message`）字段
- 每种具体事件类型 MUST 可序列化为 JSON 并可从 JSON 反序列化（round-trip 一致）
- 项目 MUST 定义事件类型常量字符串，每种事件类型对应一个唯一常量
- 项目 MUST 在 `internal/core/` 包中定义 `EventEmitter` 接口，包含 `Emit(event Event)` 方法
- 项目 MUST 在 `internal/core/` 包中定义 `EventSubscriber` 接口，包含 `Subscribe(eventType string, handler func(Event))` 方法
- `EventEmitter` 和 `EventSubscriber` 接口 MUST 可由外部包实现
- `pkg/sdk/` 和 `api/proto/` 包 MUST NOT 被此变更修改
