package agent

import (
	"encoding/json"
	"sync"
	"testing"

	"github.com/kw12121212/spec-coding-sdk/internal/core"
)

func TestConversation_Add(t *testing.T) {
	c := NewConversation()
	msg := NewMessage(RoleUser, "hello")
	c.Add(msg)

	if c.Len() != 1 {
		t.Fatalf("expected 1 message, got %d", c.Len())
	}

	msgs := c.Messages()
	if msgs[0].Content != "hello" {
		t.Fatalf("expected content 'hello', got %q", msgs[0].Content)
	}
}

func TestConversation_MessagesReturnsSnapshot(t *testing.T) {
	c := NewConversation()
	c.Add(NewMessage(RoleUser, "first"))
	c.Add(NewMessage(RoleAssistant, "second"))

	snapshot := c.Messages()
	if len(snapshot) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(snapshot))
	}

	// Mutating the snapshot should not affect the conversation.
	snapshot[0] = NewMessage(RoleUser, "mutated")
	original := c.Messages()
	if original[0].Content == "mutated" {
		t.Fatal("snapshot mutation affected internal state")
	}
}

func TestConversation_Len(t *testing.T) {
	c := NewConversation()
	if c.Len() != 0 {
		t.Fatalf("expected 0, got %d", c.Len())
	}
	c.Add(NewMessage(RoleUser, "a"))
	if c.Len() != 1 {
		t.Fatalf("expected 1, got %d", c.Len())
	}
	c.Add(NewMessage(RoleAssistant, "b"))
	if c.Len() != 2 {
		t.Fatalf("expected 2, got %d", c.Len())
	}
}

func TestConversation_Clear(t *testing.T) {
	c := NewConversation()
	c.Add(NewMessage(RoleUser, "hello"))
	c.Add(NewMessage(RoleAssistant, "world"))
	if c.Len() != 2 {
		t.Fatalf("expected 2, got %d", c.Len())
	}

	c.Clear()
	if c.Len() != 0 {
		t.Fatalf("expected 0 after clear, got %d", c.Len())
	}
}

func TestConversation_AddWithEmitter(t *testing.T) {
	emitter := &mockEmitter{}
	c := NewConversation(WithConversationEmitter(emitter))
	c.Add(NewMessage(RoleUser, "hello"))

	events := emitter.allEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Type != core.EventMessageAdded {
		t.Fatalf("expected type %q, got %q", core.EventMessageAdded, events[0].Type)
	}

	var payload core.MessageEvent
	if err := json.Unmarshal(events[0].Payload, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload.Role != "user" {
		t.Fatalf("expected role 'user', got %q", payload.Role)
	}
	if payload.Content != "hello" {
		t.Fatalf("expected content 'hello', got %q", payload.Content)
	}
}

func TestConversation_AddWithEmitter_TruncatesLongContent(t *testing.T) {
	emitter := &mockEmitter{}
	c := NewConversation(WithConversationEmitter(emitter))

	longContent := make([]byte, 300)
	for i := range longContent {
		longContent[i] = 'a'
	}
	c.Add(NewMessage(RoleUser, string(longContent)))

	events := emitter.allEvents()
	var payload core.MessageEvent
	_ = json.Unmarshal(events[0].Payload, &payload)
	if len(payload.Content) != 203 { // 200 + "..."
		t.Fatalf("expected 203 chars, got %d", len(payload.Content))
	}
}

func TestConversation_AddWithoutEmitter(t *testing.T) {
	c := NewConversation() // no emitter
	c.Add(NewMessage(RoleUser, "hello"))
	if c.Len() != 1 {
		t.Fatalf("expected 1 message, got %d", c.Len())
	}
}

func TestConversation_ConcurrentAdd(t *testing.T) {
	c := NewConversation()
	const goroutines = 100
	var wg sync.WaitGroup

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Add(NewMessage(RoleUser, "msg"))
		}()
	}
	wg.Wait()

	if c.Len() != goroutines {
		t.Fatalf("expected %d messages, got %d", goroutines, c.Len())
	}
}
