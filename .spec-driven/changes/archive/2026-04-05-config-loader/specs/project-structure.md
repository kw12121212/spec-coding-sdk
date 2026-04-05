# project-structure

## ADDED Requirements

### Requirement: Configuration loading

- 项目 MUST 在 `internal/core/` 包中提供 `LoadConfig(path string) (*Config, error)` 函数，读取指定路径的 YAML 文件并返回解析后的 `Config` 结构体。
- `Config` 结构体 MUST 定义在 `internal/core/` 包中，初始版本不包含具体字段，但 MUST 可被其他包引用。
- 当配置文件不存在或 YAML 格式无效时，`LoadConfig` MUST 返回非 nil 的 error，error 信息 MUST 包含文件路径或具体解析错误。
- `LoadConfig` MUST 在返回 `*Config` 前对解析结果进行验证。初始版本的验证仅确认 YAML 语法正确且文件可读取。
