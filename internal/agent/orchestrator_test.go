package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	"github.com/kw12121212/spec-coding-sdk/internal/core"
)

// mockThinker implements Thinker for testing.
type mockThinker struct {
	responses []ThinkResult
	calls     int
	mu        sync.Mutex
}

func (m *mockThinker) Think(_ context.Context, _ []Message) (ThinkResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.calls >= len(m.responses) {
		return ThinkResult{}, fmt.Errorf("mockThinker: no more responses (calls=%d)", m.calls)
	}
	r := m.responses[m.calls]
	m.calls++
	return r, nil
}

// mockRegistry implements ToolRegistry for testing.
type mockRegistry struct {
	tools map[string]core.Tool
}

func (r *mockRegistry) Get(name string) (core.Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

// helper to create a running agent with optional emitter.
func runningAgent(emitter core.EventEmitter) *BaseAgent {
	a := New(WithEmitter(emitter))
	_ = a.Start(context.Background())
	return a
}

func TestOrchestrator_SingleTurn_NoToolCall(t *testing.T) {
	thinker := &mockThinker{responses: []ThinkResult{
		{Content: "Hello! How can I help?"},
	}}
	orch := NewOrchestrator(runningAgent(nil), thinker, &mockRegistry{})
	result, err := orch.Run(context.Background(), "hi")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if result.Content != "Hello! How can I help?" {
		t.Errorf("expected content %q, got %q", "Hello! How can I help?", result.Content)
	}
	if len(result.ToolCalls) != 0 {
		t.Errorf("expected no tool calls, got %d", len(result.ToolCalls))
	}
}

func TestOrchestrator_SingleToolCall(t *testing.T) {
	thinker := &mockThinker{responses: []ThinkResult{
		{Content: "", ToolCalls: []ToolCall{{Name: "echo", Input: json.RawMessage(`"hello"`)}}},
		{Content: "Echo result: hello"},
	}}
	reg := &mockRegistry{tools: map[string]core.Tool{
		"echo": &stubTool{result: core.ToolResult{Output: "hello"}},
	}}
	orch := NewOrchestrator(runningAgent(nil), thinker, reg)
	result, err := orch.Run(context.Background(), "echo hello")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if result.Content != "Echo result: hello" {
		t.Errorf("expected final content %q, got %q", "Echo result: hello", result.Content)
	}
}

func TestOrchestrator_MultipleToolCalls(t *testing.T) {
	thinker := &mockThinker{responses: []ThinkResult{
		{Content: "", ToolCalls: []ToolCall{
			{Name: "add", Input: json.RawMessage(`{"a":1,"b":2}`)},
		}},
		{Content: "", ToolCalls: []ToolCall{
			{Name: "mul", Input: json.RawMessage(`{"a":3,"b":4}`)},
		}},
		{Content: "Results: 3 and 12"},
	}}
	reg := &mockRegistry{tools: map[string]core.Tool{
		"add": &stubTool{result: core.ToolResult{Output: "3"}},
		"mul": &stubTool{result: core.ToolResult{Output: "12"}},
	}}
	orch := NewOrchestrator(runningAgent(nil), thinker, reg)
	result, err := orch.Run(context.Background(), "compute")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if result.Content != "Results: 3 and 12" {
		t.Errorf("expected final content %q, got %q", "Results: 3 and 12", result.Content)
	}
}

func TestOrchestrator_ToolError_FeedbackToLLM(t *testing.T) {
	thinker := &mockThinker{responses: []ThinkResult{
		{Content: "", ToolCalls: []ToolCall{
			{Name: "fail", Input: json.RawMessage(`{}`)},
		}},
		{Content: "I see the tool failed. Let me try something else."},
	}}
	reg := &mockRegistry{tools: map[string]core.Tool{
		"fail": &stubTool{result: core.ToolResult{Output: "permission denied", IsError: true}},
	}}
	orch := NewOrchestrator(runningAgent(nil), thinker, reg)
	result, err := orch.Run(context.Background(), "try fail")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if result.Content != "I see the tool failed. Let me try something else." {
		t.Errorf("expected fallback content, got %q", result.Content)
	}
}

func TestOrchestrator_ToolExecutionError_FeedbackToLLM(t *testing.T) {
	thinker := &mockThinker{responses: []ThinkResult{
		{Content: "", ToolCalls: []ToolCall{
			{Name: "boom", Input: json.RawMessage(`{}`)},
		}},
		{Content: "Tool crashed. Stopping."},
	}}
	reg := &mockRegistry{tools: map[string]core.Tool{
		"boom": &stubTool{err: fmt.Errorf("internal error")},
	}}
	orch := NewOrchestrator(runningAgent(nil), thinker, reg)
	result, err := orch.Run(context.Background(), "trigger")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if result.Content != "Tool crashed. Stopping." {
		t.Errorf("expected fallback content, got %q", result.Content)
	}
}

func TestOrchestrator_ToolNotFound_FeedbackToLLM(t *testing.T) {
	thinker := &mockThinker{responses: []ThinkResult{
		{Content: "", ToolCalls: []ToolCall{
			{Name: "nonexistent", Input: json.RawMessage(`{}`)},
		}},
		{Content: "Tool not available."},
	}}
	orch := NewOrchestrator(runningAgent(nil), thinker, &mockRegistry{tools: map[string]core.Tool{}})
	result, err := orch.Run(context.Background(), "use bad tool")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if result.Content != "Tool not available." {
		t.Errorf("expected fallback content, got %q", result.Content)
	}
}

func TestOrchestrator_MaxIterationsExceeded(t *testing.T) {
	// Always return a tool call → never terminates normally
	thinker := &mockThinker{responses: make([]ThinkResult, 100)}
	for i := range thinker.responses {
		thinker.responses[i] = ThinkResult{
			Content:   "",
			ToolCalls: []ToolCall{{Name: "loop", Input: json.RawMessage(`{}`)}},
		}
	}
	reg := &mockRegistry{tools: map[string]core.Tool{
		"loop": &stubTool{result: core.ToolResult{Output: "ok"}},
	}}
	orch := NewOrchestrator(runningAgent(nil), thinker, reg, WithMaxIterations(3))
	_, err := orch.Run(context.Background(), "loop forever")
	if err == nil {
		t.Fatal("expected error for max iterations")
	}
	if err.Error() != "max iterations (3) reached" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestOrchestrator_NotRunning(t *testing.T) {
	agent := New() // StateInit, not started
	thinker := &mockThinker{responses: []ThinkResult{{Content: "hi"}}}
	orch := NewOrchestrator(agent, thinker, &mockRegistry{})
	_, err := orch.Run(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error when agent not running")
	}
}

func TestOrchestrator_EventSequence(t *testing.T) {
	em := &mockEmitter{}
	thinker := &mockThinker{responses: []ThinkResult{
		{Content: "", ToolCalls: []ToolCall{
			{Name: "echo", Input: json.RawMessage(`"hi"`)},
		}},
		{Content: "done"},
	}}
	reg := &mockRegistry{tools: map[string]core.Tool{
		"echo": &stubTool{result: core.ToolResult{Output: "hi"}},
	}}
	orch := NewOrchestrator(runningAgent(em), thinker, reg)
	_, err := orch.Run(context.Background(), "test events")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	events := em.allEvents()
	// Filter to orchestrator events only (Conversation also emits message.added)
	var orchEvents []core.Event
	for _, e := range events {
		switch e.Type {
		case core.EventOrchestratorThink, core.EventOrchestratorAct,
			core.EventOrchestratorObserve, core.EventOrchestratorComplete:
			orchEvents = append(orchEvents, e)
		}
	}

	// Expected: think(1), act(1,echo), observe(1), think(2), complete(2)
	expectedTypes := []string{
		core.EventOrchestratorThink,
		core.EventOrchestratorAct,
		core.EventOrchestratorObserve,
		core.EventOrchestratorThink,
		core.EventOrchestratorComplete,
	}
	if len(orchEvents) != len(expectedTypes) {
		t.Fatalf("expected %d events, got %d", len(expectedTypes), len(orchEvents))
	}
	for i, exp := range expectedTypes {
		if orchEvents[i].Type != exp {
			t.Errorf("event %d: expected type %q, got %q", i, exp, orchEvents[i].Type)
		}
	}

	// Verify act event details
	var actEvent core.OrchestratorActEvent
	if err := json.Unmarshal(orchEvents[1].Payload, &actEvent); err != nil {
		t.Fatalf("unmarshal act event: %v", err)
	}
	if actEvent.ToolName != "echo" || !actEvent.Success || actEvent.Iteration != 1 {
		t.Errorf("unexpected act event: %+v", actEvent)
	}

	// Verify complete event details
	var completeEvent core.OrchestratorCompleteEvent
	if err := json.Unmarshal(orchEvents[4].Payload, &completeEvent); err != nil {
		t.Fatalf("unmarshal complete event: %v", err)
	}
	if completeEvent.TotalIterations != 2 || completeEvent.ToolCallsMade != 1 {
		t.Errorf("unexpected complete event: %+v", completeEvent)
	}
}

func TestOrchestrator_WithMaxIterations(t *testing.T) {
	orch := NewOrchestrator(nil, nil, nil, WithMaxIterations(10))
	if orch.maxIterations != 10 {
		t.Errorf("expected maxIterations=10, got %d", orch.maxIterations)
	}
}

func TestOrchestrator_DefaultMaxIterations(t *testing.T) {
	orch := NewOrchestrator(nil, nil, nil)
	if orch.maxIterations != 50 {
		t.Errorf("expected default maxIterations=50, got %d", orch.maxIterations)
	}
}

func TestOrchestrator_ContextCancellation(t *testing.T) {
	thinker := &mockThinker{responses: make([]ThinkResult, 100)}
	for i := range thinker.responses {
		thinker.responses[i] = ThinkResult{
			Content:   "",
			ToolCalls: []ToolCall{{Name: "loop", Input: json.RawMessage(`{}`)}},
		}
	}
	reg := &mockRegistry{tools: map[string]core.Tool{
		"loop": &stubTool{result: core.ToolResult{Output: "ok"}},
	}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	orch := NewOrchestrator(runningAgent(nil), thinker, reg)
	_, err := orch.Run(ctx, "cancelled")
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

func TestOrchestrator_ConversationContainsAllMessages(t *testing.T) {
	agent := runningAgent(nil)
	thinker := &mockThinker{responses: []ThinkResult{
		{Content: "", ToolCalls: []ToolCall{
			{Name: "echo", Input: json.RawMessage(`"hello"`)},
		}},
		{Content: "final answer"},
	}}
	reg := &mockRegistry{tools: map[string]core.Tool{
		"echo": &stubTool{result: core.ToolResult{Output: "hello"}},
	}}
	orch := NewOrchestrator(agent, thinker, reg)
	_, err := orch.Run(context.Background(), "test")
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	msgs := agent.Conversation().Messages()
	// Expected: user, assistant(tool call), tool(echo), assistant(final)
	if len(msgs) != 4 {
		t.Fatalf("expected 4 messages, got %d", len(msgs))
	}
	if msgs[0].Role != RoleUser {
		t.Errorf("msg 0: expected role user, got %s", msgs[0].Role)
	}
	if msgs[1].Role != RoleAssistant {
		t.Errorf("msg 1: expected role assistant, got %s", msgs[1].Role)
	}
	if msgs[2].Role != RoleTool || msgs[2].ToolName != "echo" {
		t.Errorf("msg 2: expected tool/echo, got %s/%s", msgs[2].Role, msgs[2].ToolName)
	}
	if msgs[3].Role != RoleAssistant || msgs[3].Content != "final answer" {
		t.Errorf("msg 3: expected assistant/final answer, got %s/%s", msgs[3].Role, msgs[3].Content)
	}
}
