package agent

import (
	"time"
)

// Role represents the role of a message sender in a conversation.
type Role string

const (
	// RoleUser represents a message from the user.
	RoleUser Role = "user"
	// RoleAssistant represents a message from the assistant (LLM).
	RoleAssistant Role = "assistant"
	// RoleTool represents a message containing a tool execution result.
	RoleTool Role = "tool"
)

// Message represents a single message in a conversation.
type Message struct {
	Role      Role       `json:"role"`
	Content   string     `json:"content"`
	ToolName  string     `json:"tool_name,omitempty"`
	Timestamp time.Time  `json:"timestamp"`
}

// NewMessage creates a new Message with the given role and content.
func NewMessage(role Role, content string) Message {
	return Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	}
}

// NewToolMessage creates a new Message with RoleTool and the given tool name and content.
func NewToolMessage(toolName, content string) Message {
	return Message{
		Role:      RoleTool,
		Content:   content,
		ToolName:  toolName,
		Timestamp: time.Now(),
	}
}
