// Package llm provides provider-agnostic types and interfaces for LLM API calls.
package llm

import (
	"context"
	"encoding/json"
)

// Role represents the role of a message in an LLM conversation.
type Role string

const (
	// RoleUser identifies a user message.
	RoleUser Role = "user"
	// RoleAssistant identifies an assistant message.
	RoleAssistant Role = "assistant"
	// RoleTool identifies a tool result message.
	RoleTool Role = "tool"
)

// ToolCall represents a single tool call requested by the LLM.
type ToolCall struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

// Message represents a single message in an LLM conversation.
type Message struct {
	Role       Role       `json:"role"`
	Content    string     `json:"content"`
	ToolName   string     `json:"tool_name,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
}

// Request represents a request to an LLM provider.
type Request struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

// Usage represents token usage statistics from an LLM response.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}

// Response represents a complete response from an LLM provider.
type Response struct {
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	Usage      Usage      `json:"usage"`
	StopReason string     `json:"stop_reason"`
}

// StreamChunk represents an incremental chunk of a streaming LLM response.
type StreamChunk struct {
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
	Usage     Usage      `json:"usage,omitzero"`
}

// StreamCallback is the function type for receiving streaming LLM response chunks.
// When the callback returns a non-nil error, the provider MUST stop streaming.
type StreamCallback func(chunk StreamChunk) error

// Provider is the provider-agnostic interface for LLM API calls.
type Provider interface {
	// Complete sends a synchronous request and returns the full response.
	Complete(ctx context.Context, req Request) (Response, error)
	// Stream sends a request and delivers response chunks via the callback.
	Stream(ctx context.Context, req Request, callback StreamCallback) error
}

// ProviderConfig holds common configuration for LLM providers.
type ProviderConfig struct {
	BaseURL string `json:"base_url"`
	APIKey  string `json:"api_key"`
	Model   string `json:"model"`
}
