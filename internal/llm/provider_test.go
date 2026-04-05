package llm

import (
	"context"
	"encoding/json"
	"testing"
)

func TestUsageJSONRoundTrip(t *testing.T) {
	original := Usage{
		PromptTokens:     100,
		CompletionTokens: 50,
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal Usage: %v", err)
	}
	var decoded Usage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal Usage: %v", err)
	}
	if decoded != original {
		t.Errorf("round-trip mismatch: got %+v, want %+v", decoded, original)
	}
}

func TestToolCallJSONRoundTrip(t *testing.T) {
	original := ToolCall{
		ID:    "call_abc123",
		Name:  "bash",
		Input: json.RawMessage(`{"command":"ls"}`),
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal ToolCall: %v", err)
	}
	var decoded ToolCall
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal ToolCall: %v", err)
	}
	if decoded.ID != original.ID || decoded.Name != original.Name {
		t.Errorf("round-trip mismatch: got %+v, want %+v", decoded, original)
	}
	if string(decoded.Input) != string(original.Input) {
		t.Errorf("input mismatch: got %s, want %s", decoded.Input, original.Input)
	}
}

func TestMessageJSONRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		msg  Message
	}{
		{
			name: "user message",
			msg:  Message{Role: RoleUser, Content: "hello"},
		},
		{
			name: "assistant with tool calls",
			msg: Message{
				Role:    RoleAssistant,
				Content: "I will run a command",
				ToolCalls: []ToolCall{
					{ID: "call_1", Name: "bash", Input: json.RawMessage(`{"command":"ls"}`)},
				},
			},
		},
		{
			name: "tool result",
			msg: Message{
				Role:       RoleTool,
				Content:    "file1.go\nfile2.go",
				ToolName:   "bash",
				ToolCallID: "call_1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.msg)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var decoded Message
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if decoded.Role != tt.msg.Role {
				t.Errorf("role mismatch: got %q, want %q", decoded.Role, tt.msg.Role)
			}
			if decoded.Content != tt.msg.Content {
				t.Errorf("content mismatch: got %q, want %q", decoded.Content, tt.msg.Content)
			}
			if decoded.ToolName != tt.msg.ToolName {
				t.Errorf("tool_name mismatch: got %q, want %q", decoded.ToolName, tt.msg.ToolName)
			}
			if decoded.ToolCallID != tt.msg.ToolCallID {
				t.Errorf("tool_call_id mismatch: got %q, want %q", decoded.ToolCallID, tt.msg.ToolCallID)
			}
			if len(decoded.ToolCalls) != len(tt.msg.ToolCalls) {
				t.Fatalf("tool_calls length mismatch: got %d, want %d", len(decoded.ToolCalls), len(tt.msg.ToolCalls))
			}
			for i, tc := range decoded.ToolCalls {
				if tc.ID != tt.msg.ToolCalls[i].ID || tc.Name != tt.msg.ToolCalls[i].Name {
					t.Errorf("tool_calls[%d] mismatch: got %+v, want %+v", i, tc, tt.msg.ToolCalls[i])
				}
			}
		})
	}
}

func TestRequestJSONRoundTrip(t *testing.T) {
	original := Request{
		Model:       "gpt-4",
		Temperature: 0.7,
		MaxTokens:   1024,
		Messages: []Message{
			{Role: RoleUser, Content: "hello"},
			{Role: RoleAssistant, Content: "hi"},
		},
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal Request: %v", err)
	}
	var decoded Request
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal Request: %v", err)
	}
	if decoded.Model != original.Model {
		t.Errorf("model mismatch: got %q, want %q", decoded.Model, original.Model)
	}
	if decoded.Temperature != original.Temperature {
		t.Errorf("temperature mismatch: got %f, want %f", decoded.Temperature, original.Temperature)
	}
	if decoded.MaxTokens != original.MaxTokens {
		t.Errorf("max_tokens mismatch: got %d, want %d", decoded.MaxTokens, original.MaxTokens)
	}
	if len(decoded.Messages) != len(original.Messages) {
		t.Fatalf("messages length mismatch: got %d, want %d", len(decoded.Messages), len(original.Messages))
	}
}

func TestResponseJSONRoundTrip(t *testing.T) {
	original := Response{
		Content:    "Here is the answer",
		StopReason: "stop",
		Usage:      Usage{PromptTokens: 50, CompletionTokens: 30},
		ToolCalls: []ToolCall{
			{ID: "call_1", Name: "bash", Input: json.RawMessage(`{"command":"pwd"}`)},
		},
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal Response: %v", err)
	}
	var decoded Response
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal Response: %v", err)
	}
	if decoded.Content != original.Content {
		t.Errorf("content mismatch: got %q, want %q", decoded.Content, original.Content)
	}
	if decoded.StopReason != original.StopReason {
		t.Errorf("stop_reason mismatch: got %q, want %q", decoded.StopReason, original.StopReason)
	}
	if decoded.Usage != original.Usage {
		t.Errorf("usage mismatch: got %+v, want %+v", decoded.Usage, original.Usage)
	}
	if len(decoded.ToolCalls) != 1 || decoded.ToolCalls[0].Name != "bash" {
		t.Errorf("tool_calls mismatch: got %+v", decoded.ToolCalls)
	}
}

// mockProvider verifies that the Provider interface can be implemented by external packages.
type mockProvider struct{}

func (mockProvider) Complete(_ context.Context, _ Request) (Response, error) {
	return Response{}, nil
}

func (mockProvider) Stream(_ context.Context, _ Request, _ StreamCallback) error {
	return nil
}

// Compile-time check that mockProvider satisfies Provider.
var _ Provider = mockProvider{}

func TestProviderInterfaceImplemented(t *testing.T) {
	var p Provider = mockProvider{}
	resp, err := p.Complete(context.Background(), Request{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != "" {
		t.Errorf("expected empty content from mock, got %q", resp.Content)
	}
}
