// Package agent provides the base agent implementation with lifecycle management.
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/kw12121212/spec-coding-sdk/internal/core"
)

// State represents the lifecycle state of an agent.
type State int

const (
	// StateInit is the initial agent state before Start is called.
	StateInit State = iota
	// StateRunning means the agent is active and can execute tools.
	StateRunning
	// StatePaused means the agent is temporarily suspended.
	StatePaused
	// StateStopped means the agent has been shut down.
	StateStopped
	// StateError means the agent encountered a fatal error.
	StateError
)

func (s State) String() string {
	switch s {
	case StateInit:
		return "init"
	case StateRunning:
		return "running"
	case StatePaused:
		return "paused"
	case StateStopped:
		return "stopped"
	case StateError:
		return "error"
	default:
		return "unknown"
	}
}

// transitions defines the allowed state transitions: from-state → set of valid to-states.
var transitions = map[State]map[State]bool{
	StateInit:    {StateRunning: true},
	StateRunning: {StatePaused: true, StateStopped: true, StateError: true},
	StatePaused:  {StateRunning: true, StateStopped: true, StateError: true},
	StateError:   {StateStopped: true, StateError: true},
	StateStopped: {StateError: true},
}

// canTransition reports whether a transition from one state to another is allowed.
func canTransition(from, to State) bool {
	allowed, ok := transitions[from]
	if !ok {
		return false
	}
	return allowed[to]
}

// Option is a functional option for configuring a BaseAgent.
type Option func(*BaseAgent)

// WithEmitter sets the EventEmitter for the agent.
func WithEmitter(emitter core.EventEmitter) Option {
	return func(a *BaseAgent) {
		a.emitter = emitter
	}
}

// BaseAgent implements core.Agent with lifecycle state management.
type BaseAgent struct {
	mu      sync.RWMutex
	state   State
	emitter core.EventEmitter
}

// Compile-time check that BaseAgent satisfies core.Agent.
var _ core.Agent = (*BaseAgent)(nil)

// New creates a new BaseAgent with the given options.
func New(opts ...Option) *BaseAgent {
	a := &BaseAgent{
		state: StateInit,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// State returns the current agent state. It is safe to call from multiple goroutines.
func (a *BaseAgent) State() State {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.state
}

// Start transitions the agent from init to running.
func (a *BaseAgent) Start(_ context.Context) error {
	return a.transition(StateInit, StateRunning, "agent started")
}

// Stop transitions the agent to stopped from running, paused, or error.
func (a *BaseAgent) Stop(_ context.Context) error {
	current := a.State()
	switch current {
	case StateRunning, StatePaused, StateError:
		return a.transition(current, StateStopped, "agent stopped")
	default:
		return fmt.Errorf("cannot stop agent in state %s", current)
	}
}

// Pause transitions the agent from running to paused.
func (a *BaseAgent) Pause(_ context.Context) error {
	return a.transition(StateRunning, StatePaused, "agent paused")
}

// Resume transitions the agent from paused to running.
func (a *BaseAgent) Resume(_ context.Context) error {
	return a.transition(StatePaused, StateRunning, "agent resumed")
}

// RunTool executes a tool if the agent is in the running state.
func (a *BaseAgent) RunTool(ctx context.Context, tool core.Tool, input json.RawMessage) (core.ToolResult, error) {
	a.mu.RLock()
	current := a.state
	a.mu.RUnlock()

	if current != StateRunning {
		return core.ToolResult{
			Output:  fmt.Sprintf("agent not running (state: %s)", current),
			IsError: true,
		}, nil
	}

	return tool.Execute(ctx, input)
}

// setError transitions the agent to the error state from any state.
// This is an internal method for use by higher-level orchestration.
func (a *BaseAgent) setError(msg string) {
	_ = a.transition(a.State(), StateError, msg)
}

// transition performs a state transition with validation and event emission.
func (a *BaseAgent) transition(from, to State, msg string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.state != from {
		return fmt.Errorf("invalid transition: expected state %s but agent is in state %s", from, a.state)
	}

	if !canTransition(from, to) {
		return fmt.Errorf("transition from %s to %s is not allowed", from, to)
	}

	a.state = to
	if a.emitter != nil {
		payload, _ := json.Marshal(core.AgentStateEvent{
			State:   to.String(),
			Message: msg,
		})
		a.emitter.Emit(core.Event{
			Type:      core.EventAgentState,
			Payload:   payload,
			Timestamp: time.Now(),
		})
	}
	return nil
}
