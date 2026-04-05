package core

// Event type constants identify each kind of structured event.
const (
	EventToolCall    = "tool.call"
	EventToolResult  = "tool.result"
	EventAgentState  = "agent.state"
	EventError       = "error"
	EventMessageAdded = "message.added"
)

// ToolCallEvent is emitted when a tool is invoked.
type ToolCallEvent struct {
	ToolName string `json:"tool_name"`
	Input    string `json:"input"`
}

// EventType returns the event type constant for ToolCallEvent.
func (ToolCallEvent) EventType() string { return EventToolCall }

// ToolResultEvent is emitted when a tool execution completes.
type ToolResultEvent struct {
	ToolName string `json:"tool_name"`
	Result   string `json:"result"`
}

// EventType returns the event type constant for ToolResultEvent.
func (ToolResultEvent) EventType() string { return EventToolResult }

// AgentStateEvent is emitted when an agent's lifecycle state changes.
type AgentStateEvent struct {
	State   string `json:"state"`
	Message string `json:"message"`
}

// EventType returns the event type constant for AgentStateEvent.
func (AgentStateEvent) EventType() string { return EventAgentState }

// ErrorEvent is emitted when an error occurs during agent or tool execution.
type ErrorEvent struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// EventType returns the event type constant for ErrorEvent.
func (ErrorEvent) EventType() string { return EventError }

// MessageEvent is emitted when a message is added to a conversation.
type MessageEvent struct {
	Role     string `json:"role"`
	Content  string `json:"content"`
	ToolName string `json:"tool_name,omitempty"`
}

// EventType returns the event type constant for MessageEvent.
func (MessageEvent) EventType() string { return EventMessageAdded }

// EventEmitter is the interface for emitting structured events.
type EventEmitter interface {
	Emit(event Event)
}

// EventSubscriber is the interface for subscribing to events by type.
type EventSubscriber interface {
	Subscribe(eventType string, handler func(Event))
}
