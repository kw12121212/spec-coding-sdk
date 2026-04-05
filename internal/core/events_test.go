package core_test

import (
	"encoding/json"
	"testing"

	"github.com/kw12121212/spec-coding-sdk/internal/core"
)

func TestToolCallEventFields(t *testing.T) {
	e := core.ToolCallEvent{ToolName: "bash", Input: "ls -la"}
	if e.ToolName != "bash" {
		t.Errorf("ToolName = %q, want %q", e.ToolName, "bash")
	}
	if e.Input != "ls -la" {
		t.Errorf("Input = %q, want %q", e.Input, "ls -la")
	}
	if e.EventType() != core.EventToolCall {
		t.Errorf("EventType() = %q, want %q", e.EventType(), core.EventToolCall)
	}
}

func TestToolResultEventFields(t *testing.T) {
	e := core.ToolResultEvent{ToolName: "bash", Result: "file.txt"}
	if e.ToolName != "bash" {
		t.Errorf("ToolName = %q, want %q", e.ToolName, "bash")
	}
	if e.Result != "file.txt" {
		t.Errorf("Result = %q, want %q", e.Result, "file.txt")
	}
	if e.EventType() != core.EventToolResult {
		t.Errorf("EventType() = %q, want %q", e.EventType(), core.EventToolResult)
	}
}

func TestAgentStateEventFields(t *testing.T) {
	e := core.AgentStateEvent{State: "running", Message: "started"}
	if e.State != "running" {
		t.Errorf("State = %q, want %q", e.State, "running")
	}
	if e.Message != "started" {
		t.Errorf("Message = %q, want %q", e.Message, "started")
	}
	if e.EventType() != core.EventAgentState {
		t.Errorf("EventType() = %q, want %q", e.EventType(), core.EventAgentState)
	}
}

func TestErrorEventFields(t *testing.T) {
	e := core.ErrorEvent{Code: "E001", Message: "something failed"}
	if e.Code != "E001" {
		t.Errorf("Code = %q, want %q", e.Code, "E001")
	}
	if e.Message != "something failed" {
		t.Errorf("Message = %q, want %q", e.Message, "something failed")
	}
	if e.EventType() != core.EventError {
		t.Errorf("EventType() = %q, want %q", e.EventType(), core.EventError)
	}
}

func TestEventConstantsUnique(t *testing.T) {
	consts := map[string]string{
		core.EventToolCall:   "EventToolCall",
		core.EventToolResult: "EventToolResult",
		core.EventAgentState: "EventAgentState",
		core.EventError:      "EventError",
	}
	if len(consts) != 4 {
		t.Errorf("expected 4 unique constants, got %d", len(consts))
	}
}

func TestToolCallEventJSONRoundTrip(t *testing.T) {
	original := core.ToolCallEvent{ToolName: "grep", Input: `"pattern"`}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var decoded core.ToolCallEvent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded != original {
		t.Errorf("round-trip mismatch: got %+v, want %+v", decoded, original)
	}
}

func TestToolResultEventJSONRoundTrip(t *testing.T) {
	original := core.ToolResultEvent{ToolName: "grep", Result: "match found"}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var decoded core.ToolResultEvent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded != original {
		t.Errorf("round-trip mismatch: got %+v, want %+v", decoded, original)
	}
}

func TestAgentStateEventJSONRoundTrip(t *testing.T) {
	original := core.AgentStateEvent{State: "idle", Message: "waiting for input"}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var decoded core.AgentStateEvent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded != original {
		t.Errorf("round-trip mismatch: got %+v, want %+v", decoded, original)
	}
}

func TestErrorEventJSONRoundTrip(t *testing.T) {
	original := core.ErrorEvent{Code: "E500", Message: "internal error"}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var decoded core.ErrorEvent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded != original {
		t.Errorf("round-trip mismatch: got %+v, want %+v", decoded, original)
	}
}

// stubEmitter satisfies EventEmitter from outside the core package.
type stubEmitter struct{ last core.Event }

func (s *stubEmitter) Emit(event core.Event) { s.last = event }

func TestEventEmitterImplemented(t *testing.T) {
	var _ core.EventEmitter = &stubEmitter{}
	emitter := &stubEmitter{}
	evt := core.Event{Type: core.EventToolCall}
	emitter.Emit(evt)
	if emitter.last.Type != core.EventToolCall {
		t.Errorf("Emit did not store event: got type %q", emitter.last.Type)
	}
}

// stubSubscriber satisfies EventSubscriber from outside the core package.
type stubSubscriber struct {
	called bool
}

func (s *stubSubscriber) Subscribe(_ string, _ func(core.Event)) {
	s.called = true
}

func TestEventSubscriberImplemented(t *testing.T) {
	var _ core.EventSubscriber = &stubSubscriber{}
	sub := &stubSubscriber{}
	sub.Subscribe(core.EventToolCall, func(core.Event) {})
	if !sub.called {
		t.Error("Subscribe was not called")
	}
}
