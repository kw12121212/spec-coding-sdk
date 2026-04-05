# M15 - auto-spec-driven 集成与收尾

## Goal

集成内置 auto-spec-driven skill，完成发布准备。

## In Scope

- auto-spec-driven skill 内置集成（来自 ../auto-spec-driven）
- 发布准备（文档、CHANGELOG、发布脚本）

## Out of Scope

- auto-spec-driven skill 本身的开发（外部依赖）
- CI/CD pipeline 搭建（可作为后续增强）
- 各接口层的端到端测试（已在 M12-M14 各自完成）

## Done Criteria

- auto-spec-driven skill 可通过 SDK 接口被正确调用
- README 和 API 文档完整
- 可执行一次完整的 demo 流程
- 发布脚本支持多平台交叉编译

## Planned Changes

- `auto-spec-driven-integration` - 内置 skill 集成与接口层适配
- `release-prep` - 文档完善与发布脚本实现

## Dependencies

- M11 Native Go SDK 层
- auto-spec-driven skill 代码可用

## Risks

- auto-spec-driven skill 的接口变更可能影响集成

## Status

- Declared: proposed

## Notes

- 此里程碑是整个项目的收尾阶段，各接口层的行为一致性已在 M12-M14 各自的 e2e 测试中验证
- 发布脚本需支持多平台交叉编译（linux/darwin/windows, amd64/arm64）
