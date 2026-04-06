# permissions

## ADDED Requirements

### Requirement: 非交互执行路径中的确认决策

- 在仅通过 `core.PermissionProvider.Check(ctx, operation, resource) error` 暴露权限结果、且未提供人工确认通道的执行路径中，项目 MUST 将 `DecisionNeedConfirmation` 视为阻断执行，而不是隐式允许。
- 当权限检查结果为拒绝或需确认时，项目 MUST 在产生任何工具副作用之前停止该次工具调用。
- 对基于 `core.PermissionProvider` 的工具执行路径，阻断结果 MUST 通过 `ToolResult{IsError: true}` 返回给调用方，且 `Output` MUST 保留 `PermissionProvider.Check` 返回的错误消息，以便调用方区分拒绝与需确认。
- 项目 MUST 包含测试，验证至少一个会启动进程的工具和一个会修改文件系统的工具在拒绝或需确认时不会产生副作用。
