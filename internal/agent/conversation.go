package agent

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/kw12121212/spec-coding-sdk/internal/core"
)

// ConversationOption is a functional option for configuring a Conversation.
type ConversationOption func(*Conversation)

// WithConversationEmitter sets the EventEmitter for the conversation.
func WithConversationEmitter(emitter core.EventEmitter) ConversationOption {
	return func(c *Conversation) {
		c.emitter = emitter
	}
}

// Conversation manages an ordered list of messages for an agent session.
type Conversation struct {
	mu       sync.RWMutex
	messages []Message
	emitter  core.EventEmitter
}

// NewConversation creates a new Conversation with the given options.
func NewConversation(opts ...ConversationOption) *Conversation {
	c := &Conversation{}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

const contentPreviewLimit = 200

// Add appends a message to the conversation and emits a message.added event.
func (c *Conversation) Add(msg Message) {
	c.mu.Lock()
	c.messages = append(c.messages, msg)
	c.mu.Unlock()

	if c.emitter != nil {
		preview := msg.Content
		if len(preview) > contentPreviewLimit {
			preview = preview[:contentPreviewLimit] + "..."
		}
		payload, _ := json.Marshal(core.MessageEvent{
			Role:     string(msg.Role),
			Content:  preview,
			ToolName: msg.ToolName,
		})
		c.emitter.Emit(core.Event{
			Type:      core.EventMessageAdded,
			Payload:   payload,
			Timestamp: time.Now(),
		})
	}
}

// Messages returns a snapshot (copy) of the current message list.
func (c *Conversation) Messages() []Message {
	c.mu.RLock()
	defer c.mu.RUnlock()
	snapshot := make([]Message, len(c.messages))
	copy(snapshot, c.messages)
	return snapshot
}

// Len returns the number of messages in the conversation.
func (c *Conversation) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.messages)
}

// Clear removes all messages from the conversation.
func (c *Conversation) Clear() {
	c.mu.Lock()
	c.messages = nil
	c.mu.Unlock()
}
