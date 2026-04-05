// Package core provides internal core interfaces and types for spec-coding-sdk.
package core

import (
	"context"
	"encoding/json"
	"time"
)

// ToolResult holds the result of a tool execution.
type ToolResult struct {
	Output  string `json:"output"`
	IsError bool   `json:"is_error"`
}

// Tool is the contract every tool must satisfy.
type Tool interface {
	Execute(ctx context.Context, input json.RawMessage) (ToolResult, error)
}

// Agent is the agent lifecycle contract covering session management and tool invocation.
type Agent interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	RunTool(ctx context.Context, tool Tool, input json.RawMessage) (ToolResult, error)
}

// Event represents a structured event in the event system.
type Event struct {
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp time.Time       `json:"timestamp"`
}

// PermissionProvider is the permission check contract for tool execution.
type PermissionProvider interface {
	Check(ctx context.Context, operation string, resource string) error
}
