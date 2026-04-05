package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kw12121212/spec-coding-sdk/internal/core"
)

// ToolCall represents a single tool call request returned by the LLM.
type ToolCall struct {
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

// ThinkResult represents the LLM's response, which may include tool calls or a final text reply.
type ThinkResult struct {
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls"`
}

// Thinker is the interface for LLM-based thinking. Implementations receive the
// current message list and return either a final reply or tool call requests.
type Thinker interface {
	Think(ctx context.Context, messages []Message) (ThinkResult, error)
}

// ToolRegistry looks up tools by name for the orchestrator.
type ToolRegistry interface {
	Get(name string) (core.Tool, bool)
}

// OrchestratorOption configures an Orchestrator.
type OrchestratorOption func(*Orchestrator)

const defaultMaxIterations = 50

// WithMaxIterations sets the maximum number of think-act-observe iterations.
func WithMaxIterations(n int) OrchestratorOption {
	return func(o *Orchestrator) {
		o.maxIterations = n
	}
}

// Orchestrator drives the think-act-observe loop for a BaseAgent.
type Orchestrator struct {
	agent          *BaseAgent
	thinker        Thinker
	registry       ToolRegistry
	maxIterations  int
}

// NewOrchestrator creates a new Orchestrator.
func NewOrchestrator(agent *BaseAgent, thinker Thinker, registry ToolRegistry, opts ...OrchestratorOption) *Orchestrator {
	o := &Orchestrator{
		agent:         agent,
		thinker:       thinker,
		registry:      registry,
		maxIterations: defaultMaxIterations,
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// Run executes the full orchestration loop: add user message, then think-act-observe
// until the LLM produces a final reply or the iteration limit is reached.
func (o *Orchestrator) Run(ctx context.Context, userMessage string) (ThinkResult, error) {
	if o.agent.State() != StateRunning {
		return ThinkResult{}, fmt.Errorf("agent not running (state: %s)", o.agent.State())
	}

	conv := o.agent.Conversation()
	conv.Add(NewMessage(RoleUser, userMessage))

	var (
		totalToolCalls int
		result         ThinkResult
	)

	for i := 1; i <= o.maxIterations; i++ {
		if err := ctx.Err(); err != nil {
			return ThinkResult{}, err
		}

		// Think
		o.emitEvent(core.OrchestratorThinkEvent{Iteration: i})
		var err error
		result, err = o.thinker.Think(ctx, conv.Messages())
		if err != nil {
			return ThinkResult{}, fmt.Errorf("think failed at iteration %d: %w", i, err)
		}

		// No tool calls → final reply
		if len(result.ToolCalls) == 0 {
			conv.Add(NewMessage(RoleAssistant, result.Content))
			o.emitEvent(core.OrchestratorCompleteEvent{
				TotalIterations: i,
				ToolCallsMade:   totalToolCalls,
			})
			return result, nil
		}

		// Act: add assistant message with tool calls summary, then execute each tool
		conv.Add(NewMessage(RoleAssistant, result.Content))

		for _, tc := range result.ToolCalls {
			tool, found := o.registry.Get(tc.Name)
			if !found {
				totalToolCalls++
				o.emitEvent(core.OrchestratorActEvent{
					Iteration: i,
					ToolName:  tc.Name,
					Success:   false,
				})
				conv.Add(NewToolMessage(tc.Name, fmt.Sprintf("tool not found: %s", tc.Name)))
				o.emitEvent(core.OrchestratorObserveEvent{
					Iteration:   i,
					MessageCount: conv.Len(),
				})
				continue
			}

			toolResult, execErr := tool.Execute(ctx, tc.Input)
			totalToolCalls++
			success := execErr == nil && !toolResult.IsError
			o.emitEvent(core.OrchestratorActEvent{
				Iteration: i,
				ToolName:  tc.Name,
				Success:   success,
			})

			content := toolResult.Output
			if execErr != nil {
				content = fmt.Sprintf("tool execution error: %s", execErr)
			}
			conv.Add(NewToolMessage(tc.Name, content))
			o.emitEvent(core.OrchestratorObserveEvent{
				Iteration:    i,
				MessageCount: conv.Len(),
			})
		}
	}

	return ThinkResult{}, fmt.Errorf("max iterations (%d) reached", o.maxIterations)
}

func (o *Orchestrator) emitEvent(ev interface{ EventType() string }) {
	emitter := o.agent.emitter
	if emitter == nil {
		return
	}
	payload, _ := json.Marshal(ev)
	emitter.Emit(core.Event{
		Type:      ev.EventType(),
		Payload:   payload,
		Timestamp: time.Now(),
	})
}
