package agent

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	"github.com/kw12121212/spec-coding-sdk/internal/core"
)

// stubTool is a minimal Tool implementation for testing.
type stubTool struct {
	result core.ToolResult
	err    error
}

func (t *stubTool) Execute(_ context.Context, _ json.RawMessage) (core.ToolResult, error) {
	return t.result, t.err
}

// mockEmitter captures emitted events.
type mockEmitter struct {
	mu     sync.Mutex
	events []core.Event
}

func (m *mockEmitter) Emit(event core.Event) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, event)
}

func (m *mockEmitter) allEvents() []core.Event {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]core.Event, len(m.events))
	copy(out, m.events)
	return out
}

func TestNew_InitialState(t *testing.T) {
	a := New()
	if a.State() != StateInit {
		t.Fatalf("expected initial state %s, got %s", StateInit, a.State())
	}
}

func TestStart_Success(t *testing.T) {
	emitter := &mockEmitter{}
	a := New(WithEmitter(emitter))

	err := a.Start(context.Background())
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	if a.State() != StateRunning {
		t.Fatalf("expected state %s, got %s", StateRunning, a.State())
	}

	events := emitter.allEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	var evt core.AgentStateEvent
	if err := json.Unmarshal(events[0].Payload, &evt); err != nil {
		t.Fatalf("unmarshal event payload: %v", err)
	}
	if evt.State != "running" {
		t.Fatalf("expected event state 'running', got %q", evt.State)
	}
	if evt.Message != "agent started" {
		t.Fatalf("expected event message 'agent started', got %q", evt.Message)
	}
}

func TestStart_InvalidState(t *testing.T) {
	tests := []struct {
		name  string
		setup func(a *BaseAgent)
	}{
		{"already running", func(a *BaseAgent) { _ = a.Start(context.Background()) }},
		{"stopped", func(a *BaseAgent) {
			_ = a.Start(context.Background())
			_ = a.Stop(context.Background())
		}},
		{"paused", func(a *BaseAgent) {
			_ = a.Start(context.Background())
			_ = a.Pause(context.Background())
		}},
		{"error state", func(a *BaseAgent) {
			_ = a.Start(context.Background())
			a.setError("test")
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := New()
			tt.setup(a)
			err := a.Start(context.Background())
			if err == nil {
				t.Fatal("expected error when starting from invalid state")
			}
		})
	}
}

func TestStop_Success(t *testing.T) {
	tests := []struct {
		name  string
		setup func(a *BaseAgent)
	}{
		{"from running", func(a *BaseAgent) { _ = a.Start(context.Background()) }},
		{"from paused", func(a *BaseAgent) {
			_ = a.Start(context.Background())
			_ = a.Pause(context.Background())
		}},
		{"from error", func(a *BaseAgent) {
			_ = a.Start(context.Background())
			a.setError("test")
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emitter := &mockEmitter{}
			a := New(WithEmitter(emitter))
			tt.setup(a)

			err := a.Stop(context.Background())
			if err != nil {
				t.Fatalf("Stop failed: %v", err)
			}
			if a.State() != StateStopped {
				t.Fatalf("expected state %s, got %s", StateStopped, a.State())
			}

			events := emitter.allEvents()
			var lastEvent core.AgentStateEvent
			if err := json.Unmarshal(events[len(events)-1].Payload, &lastEvent); err != nil {
				t.Fatalf("unmarshal event payload: %v", err)
			}
			if lastEvent.State != "stopped" {
				t.Fatalf("expected event state 'stopped', got %q", lastEvent.State)
			}
		})
	}
}

func TestStop_InvalidState(t *testing.T) {
	tests := []struct {
		name  string
		setup func(a *BaseAgent)
	}{
		{"init", func(*BaseAgent) {}},
		{"already stopped", func(a *BaseAgent) {
			_ = a.Start(context.Background())
			_ = a.Stop(context.Background())
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := New()
			tt.setup(a)
			err := a.Stop(context.Background())
			if err == nil {
				t.Fatal("expected error when stopping from invalid state")
			}
		})
	}
}

func TestPause_Success(t *testing.T) {
	emitter := &mockEmitter{}
	a := New(WithEmitter(emitter))
	_ = a.Start(context.Background())

	err := a.Pause(context.Background())
	if err != nil {
		t.Fatalf("Pause failed: %v", err)
	}
	if a.State() != StatePaused {
		t.Fatalf("expected state %s, got %s", StatePaused, a.State())
	}

	events := emitter.allEvents()
	var lastEvent core.AgentStateEvent
	if err := json.Unmarshal(events[len(events)-1].Payload, &lastEvent); err != nil {
		t.Fatalf("unmarshal event payload: %v", err)
	}
	if lastEvent.State != "paused" {
		t.Fatalf("expected event state 'paused', got %q", lastEvent.State)
	}
}

func TestPause_InvalidState(t *testing.T) {
	tests := []struct {
		name  string
		setup func(a *BaseAgent)
	}{
		{"init", func(*BaseAgent) {}},
		{"paused", func(a *BaseAgent) {
			_ = a.Start(context.Background())
			_ = a.Pause(context.Background())
		}},
		{"stopped", func(a *BaseAgent) {
			_ = a.Start(context.Background())
			_ = a.Stop(context.Background())
		}},
		{"error", func(a *BaseAgent) {
			_ = a.Start(context.Background())
			a.setError("test")
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := New()
			tt.setup(a)
			err := a.Pause(context.Background())
			if err == nil {
				t.Fatal("expected error when pausing from invalid state")
			}
		})
	}
}

func TestResume_Success(t *testing.T) {
	emitter := &mockEmitter{}
	a := New(WithEmitter(emitter))
	_ = a.Start(context.Background())
	_ = a.Pause(context.Background())

	err := a.Resume(context.Background())
	if err != nil {
		t.Fatalf("Resume failed: %v", err)
	}
	if a.State() != StateRunning {
		t.Fatalf("expected state %s, got %s", StateRunning, a.State())
	}

	events := emitter.allEvents()
	var lastEvent core.AgentStateEvent
	if err := json.Unmarshal(events[len(events)-1].Payload, &lastEvent); err != nil {
		t.Fatalf("unmarshal event payload: %v", err)
	}
	if lastEvent.State != "running" {
		t.Fatalf("expected event state 'running', got %q", lastEvent.State)
	}
}

func TestResume_InvalidState(t *testing.T) {
	tests := []struct {
		name  string
		setup func(a *BaseAgent)
	}{
		{"init", func(*BaseAgent) {}},
		{"running", func(a *BaseAgent) { _ = a.Start(context.Background()) }},
		{"stopped", func(a *BaseAgent) {
			_ = a.Start(context.Background())
			_ = a.Stop(context.Background())
		}},
		{"error", func(a *BaseAgent) {
			_ = a.Start(context.Background())
			a.setError("test")
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := New()
			tt.setup(a)
			err := a.Resume(context.Background())
			if err == nil {
				t.Fatal("expected error when resuming from invalid state")
			}
		})
	}
}

func TestRunTool_RunningState(t *testing.T) {
	a := New()
	_ = a.Start(context.Background())

	tool := &stubTool{
		result: core.ToolResult{Output: "hello", IsError: false},
	}
	result, err := a.RunTool(context.Background(), tool, json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("RunTool failed: %v", err)
	}
	if result.Output != "hello" {
		t.Fatalf("expected output 'hello', got %q", result.Output)
	}
	if result.IsError {
		t.Fatal("expected IsError=false")
	}
}

func TestRunTool_NotRunning(t *testing.T) {
	tests := []struct {
		name  string
		state string
		setup func(a *BaseAgent)
	}{
		{"init", "init", func(*BaseAgent) {}},
		{"paused", "paused", func(a *BaseAgent) {
			_ = a.Start(context.Background())
			_ = a.Pause(context.Background())
		}},
		{"stopped", "stopped", func(a *BaseAgent) {
			_ = a.Start(context.Background())
			_ = a.Stop(context.Background())
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := New()
			tt.setup(a)
			result, err := a.RunTool(context.Background(), &stubTool{}, json.RawMessage(`{}`))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !result.IsError {
				t.Fatal("expected IsError=true")
			}
		})
	}
}

func TestRunTool_PropagatesToolError(t *testing.T) {
	a := New()
	_ = a.Start(context.Background())

	tool := &stubTool{
		result: core.ToolResult{Output: "tool failed", IsError: true},
	}
	result, err := a.RunTool(context.Background(), tool, json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError=true from tool")
	}
}

func TestNoEmitter_NoPanic(t *testing.T) {
	a := New() // no emitter
	_ = a.Start(context.Background())
	_ = a.Pause(context.Background())
	_ = a.Resume(context.Background())
	_ = a.Stop(context.Background())
	if a.State() != StateStopped {
		t.Fatalf("expected stopped, got %s", a.State())
	}
}

func TestEventPayload_ContainsCorrectFields(t *testing.T) {
	emitter := &mockEmitter{}
	a := New(WithEmitter(emitter))
	_ = a.Start(context.Background())

	events := emitter.allEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	evt := events[0]
	if evt.Type != core.EventAgentState {
		t.Fatalf("expected event type %q, got %q", core.EventAgentState, evt.Type)
	}
	if evt.Timestamp.IsZero() {
		t.Fatal("expected non-zero timestamp")
	}

	var payload core.AgentStateEvent
	if err := json.Unmarshal(evt.Payload, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload.State != "running" {
		t.Fatalf("expected state 'running', got %q", payload.State)
	}
	if payload.Message != "agent started" {
		t.Fatalf("expected message 'agent started', got %q", payload.Message)
	}
}

func TestConcurrentLifecycle(t *testing.T) {
	a := New()
	_ = a.Start(context.Background())

	var wg sync.WaitGroup
	const goroutines = 100

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = a.Pause(context.Background())
			_ = a.Resume(context.Background())
		}()
	}
	wg.Wait()

	// Agent should be in a valid state (running or paused are both fine due to race)
	state := a.State()
	if state != StateRunning && state != StatePaused {
		t.Fatalf("expected running or paused after concurrent operations, got %s", state)
	}
}

func TestConcurrentRunTool(t *testing.T) {
	a := New()
	_ = a.Start(context.Background())

	var wg sync.WaitGroup
	const goroutines = 50
	results := make(chan core.ToolResult, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, _ := a.RunTool(context.Background(), &stubTool{
				result: core.ToolResult{Output: "ok"},
			}, json.RawMessage(`{}`))
			results <- result
		}()
	}
	wg.Wait()
	close(results)

	count := 0
	for r := range results {
		count++
		if r.IsError {
			t.Fatalf("unexpected error result: %s", r.Output)
		}
	}
	if count != goroutines {
		t.Fatalf("expected %d results, got %d", goroutines, count)
	}
}

func TestDefaultConversation(t *testing.T) {
	a := New()
	conv := a.Conversation()
	if conv == nil {
		t.Fatal("expected non-nil default conversation")
	}
	if conv.Len() != 0 {
		t.Fatalf("expected empty conversation, got %d messages", conv.Len())
	}
}

func TestWithConversation(t *testing.T) {
	c := NewConversation()
	c.Add(NewMessage(RoleUser, "existing"))
	a := New(WithConversation(c))

	conv := a.Conversation()
	if conv.Len() != 1 {
		t.Fatalf("expected 1 message, got %d", conv.Len())
	}
	msgs := conv.Messages()
	if msgs[0].Content != "existing" {
		t.Fatalf("expected 'existing', got %q", msgs[0].Content)
	}
}

func TestSetConversation_InitState(t *testing.T) {
	a := New()
	newConv := NewConversation()
	newConv.Add(NewMessage(RoleAssistant, "new"))

	err := a.SetConversation(newConv)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Conversation().Len() != 1 {
		t.Fatalf("expected 1 message, got %d", a.Conversation().Len())
	}
}

func TestSetConversation_NonInitState(t *testing.T) {
	a := New()
	_ = a.Start(context.Background())

	err := a.SetConversation(NewConversation())
	if err == nil {
		t.Fatal("expected error when setting conversation in running state")
	}
}
