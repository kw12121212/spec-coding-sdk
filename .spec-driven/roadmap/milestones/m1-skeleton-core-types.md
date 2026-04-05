# M1 - 项目骨架与核心接口

## Goal

搭建 Go 模块、构建系统和核心接口，为后续所有功能开发提供统一的类型基础和项目结构。

## In Scope

- Go module 初始化与依赖管理
- 项目目录结构约定
- 构建、lint、测试基础设施
- 核心接口定义（Tool, Agent, Event, PermissionProvider）
- 配置加载（config.yaml 解析）
- 结构化事件系统基础类型

## Out of Scope

- 具体工具实现
- Agent 运行时逻辑
- 接口传输层（JSON-RPC / HTTP / gRPC）

## Done Criteria

- `go build ./...` 和 `go test ./...` 通过
- 核心 interface 类型可在其他包中被引用
- 配置文件可被正确解析为结构体
- 事件类型可被实例化并序列化

## Planned Changes

- `project-scaffold` - 初始化 Go module、目录结构、Makefile、lint 配置 ✅ (archived 2026-04-05)
- `core-interfaces` - 定义 Tool、Agent、Event、PermissionProvider 等核心接口 ✅ (archived 2026-04-05)
- `config-loader` - 实现 config.yaml 解析与结构体映射 ✅ (archived 2026-04-05)
- `event-system-types` - 定义结构化事件系统的核心类型

## Dependencies

- Go 1.25+ 开发环境

## Risks

- 核心接口设计需要兼顾后续多层接口（SDK / JSON-RPC / HTTP / gRPC）的需求，过早固化可能产生返工

## Status

- Declared: proposed

## Notes

- 此里程碑是所有后续开发的基础，接口设计需与 claw-code 的 Rust 实现保持**功能对等、设计对齐但不照搬**
- 建议先完成 `project-scaffold` 和 `core-interfaces` 两个 change，再推进 `config-loader` 和 `event-system-types`——核心接口的 review 不应被配置加载细节阻塞
- PermissionProvider 接口在此里程碑定义，M2 工具集可据此预留权限检查钩子
