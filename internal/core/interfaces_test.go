package core_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/kw12121212/spec-coding-sdk/internal/core"
)

// mockTool implements core.Tool for testing.
type mockTool struct{}

func (m *mockTool) Execute(_ context.Context, _ json.RawMessage) (core.ToolResult, error) {
	return core.ToolResult{Output: "mock output"}, nil
}

// mockAgent implements core.Agent for testing.
type mockAgent struct {
	started bool
}

func (m *mockAgent) Start(_ context.Context) error {
	m.started = true
	return nil
}

func (m *mockAgent) Stop(_ context.Context) error {
	m.started = false
	return nil
}

func (m *mockAgent) RunTool(_ context.Context, tool core.Tool, input json.RawMessage) (core.ToolResult, error) {
	return tool.Execute(context.Background(), input)
}

// mockPermissionProvider implements core.PermissionProvider for testing.
type mockPermissionProvider struct{}

func (m *mockPermissionProvider) Check(_ context.Context, _, _ string) error {
	return nil
}

func TestToolInterface(t *testing.T) {
	var _ core.Tool = &mockTool{}
	ctx := context.Background()
	result, err := (&mockTool{}).Execute(ctx, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Output != "mock output" {
		t.Errorf("expected 'mock output', got %q", result.Output)
	}
	if result.IsError {
		t.Error("expected IsError to be false")
	}
}

func TestAgentInterface(t *testing.T) {
	var _ core.Agent = &mockAgent{}
	ctx := context.Background()
	a := &mockAgent{}
	if err := a.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if !a.started {
		t.Error("expected agent to be started")
	}
	if err := a.Stop(ctx); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
	if a.started {
		t.Error("expected agent to be stopped")
	}
}

func TestAgentRunTool(t *testing.T) {
	a := &mockAgent{}
	tool := &mockTool{}
	input := json.RawMessage(`{"key": "value"}`)
	result, err := a.RunTool(context.Background(), tool, input)
	if err != nil {
		t.Fatalf("RunTool failed: %v", err)
	}
	if result.Output != "mock output" {
		t.Errorf("expected 'mock output', got %q", result.Output)
	}
}

func TestEventStruct(t *testing.T) {
	payload := json.RawMessage(`{"detail": "test"}`)
	now := time.Now().Truncate(time.Millisecond)
	evt := core.Event{
		Type:      "test.event",
		Payload:   payload,
		Timestamp: now,
	}
	if evt.Type != "test.event" {
		t.Errorf("expected type 'test.event', got %q", evt.Type)
	}
	if string(evt.Payload) != `{"detail": "test"}` {
		t.Errorf("unexpected payload: %q", string(evt.Payload))
	}
	if !evt.Timestamp.Equal(now) {
		t.Errorf("timestamp mismatch: expected %v, got %v", now, evt.Timestamp)
	}
}

func TestEventJSONRoundTrip(t *testing.T) {
	original := core.Event{
		Type:      "tool.executed",
		Payload:   json.RawMessage(`{"tool": "bash"}`),
		Timestamp: time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC),
	}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded core.Event
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Type != original.Type {
		t.Errorf("type mismatch: %q vs %q", decoded.Type, original.Type)
	}
	if decoded.Timestamp.UTC() != original.Timestamp {
		t.Errorf("timestamp mismatch: %v vs %v", decoded.Timestamp, original.Timestamp)
	}
}

func TestPermissionProviderInterface(t *testing.T) {
	var _ core.PermissionProvider = &mockPermissionProvider{}
	ctx := context.Background()
	pp := &mockPermissionProvider{}
	if err := pp.Check(ctx, "execute", "bash"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestToolResultType(t *testing.T) {
	tr := core.ToolResult{Output: "result", IsError: false}
	if tr.Output != "result" {
		t.Errorf("expected 'result', got %q", tr.Output)
	}
	if tr.IsError {
		t.Error("expected IsError to be false")
	}
}
