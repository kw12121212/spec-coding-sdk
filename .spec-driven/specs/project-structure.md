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
