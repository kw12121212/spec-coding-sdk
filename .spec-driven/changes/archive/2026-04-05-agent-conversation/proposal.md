# agent-conversation

## What

在 `internal/agent/` 包中实现会话与消息管理：定义消息类型（user、assistant、tool）、会话（Conversation）结构体及其消息历史管理能力，为后续编排循环（agent-orchestrator）提供对话上下文存储层。

## Why

M4 里程碑的 `agent-lifecycle` 已完成 agent 状态机与生命周期管理，但 agent 目前无法维护对话上下文。编排循环需要：
- 在多轮交互中追踪消息历史
- 区分用户消息、助手消息和工具结果消息
- 为 LLM 调用组装上下文

会话层是编排循环的必要前置。

## Scope

**In scope:**
- `Message` 类型定义（role + content 字段）
- 消息角色常量（RoleUser、RoleAssistant、RoleTool）
- `Conversation` 结构体：消息列表的增删查、上下文组装
- `Conversation` 通过 functional options 注入 EventEmitter
- 会话相关事件（message.added）
- `BaseAgent` 新增获取/设置 Conversation 的能力

**Out of scope:**
- LLM 实际调用（M5）
- 编排循环逻辑（agent-orchestrator）
- 消息持久化到磁盘/数据库
- token 计数与上下文窗口裁剪（M5）
- 权限检查（M6）

## Unchanged Behavior

- `BaseAgent` 现有的状态机（Start/Stop/Pause/Resume）行为不变
- `RunTool` 方法行为不变
- `core.Agent` 接口不变（Conversation 为 BaseAgent 扩展方法）
- 已有事件类型和结构不变
