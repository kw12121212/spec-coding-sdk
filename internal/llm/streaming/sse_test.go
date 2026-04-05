package streaming

import (
	"io"
	"strings"
	"testing"
)

func TestSSEParser_BasicEventStream(t *testing.T) {
	input := "data: hello\n\ndata: world\n\n"
	p := NewSSEParser(strings.NewReader(input))

	evt, err := p.Next()
	if err != nil {
		t.Fatalf("first event: %v", err)
	}
	if evt.Data != "hello" {
		t.Errorf("first data = %q, want %q", evt.Data, "hello")
	}

	evt, err = p.Next()
	if err != nil {
		t.Fatalf("second event: %v", err)
	}
	if evt.Data != "world" {
		t.Errorf("second data = %q, want %q", evt.Data, "world")
	}

	_, err = p.Next()
	if err != io.EOF {
		t.Errorf("after second event: got %v, want io.EOF", err)
	}
}

func TestSSEParser_WithEventType(t *testing.T) {
	input := "event: ping\ndata: pong\n\ndata: solo\n\n"
	p := NewSSEParser(strings.NewReader(input))

	evt, err := p.Next()
	if err != nil {
		t.Fatalf("first event: %v", err)
	}
	if evt.Event != "ping" {
		t.Errorf("event type = %q, want %q", evt.Event, "ping")
	}
	if evt.Data != "pong" {
		t.Errorf("data = %q, want %q", evt.Data, "pong")
	}

	evt, err = p.Next()
	if err != nil {
		t.Fatalf("second event: %v", err)
	}
	if evt.Event != "" {
		t.Errorf("second event type = %q, want empty", evt.Event)
	}
	if evt.Data != "solo" {
		t.Errorf("second data = %q, want %q", evt.Data, "solo")
	}
}

func TestSSEParser_DoneSignal(t *testing.T) {
	input := "data: chunk1\n\ndata: [DONE]\n\n"
	p := NewSSEParser(strings.NewReader(input))

	evt, err := p.Next()
	if err != nil {
		t.Fatalf("first event: %v", err)
	}
	if evt.Data != "chunk1" {
		t.Errorf("data = %q, want %q", evt.Data, "chunk1")
	}

	_, err = p.Next()
	if err != io.EOF {
		t.Errorf("after [DONE]: got %v, want io.EOF", err)
	}
	if !p.Done() {
		t.Error("Done() = false, want true")
	}
}

func TestSSEParser_MessageStop(t *testing.T) {
	input := "event: content_block_delta\ndata: {\"type\":\"text\"}\n\nevent: message_stop\n"
	p := NewSSEParser(strings.NewReader(input))

	evt, err := p.Next()
	if err != nil {
		t.Fatalf("first event: %v", err)
	}
	if evt.Event != "content_block_delta" {
		t.Errorf("event = %q, want %q", evt.Event, "content_block_delta")
	}

	_, err = p.Next()
	if err != io.EOF {
		t.Errorf("after message_stop: got %v, want io.EOF", err)
	}
	if !p.Done() {
		t.Error("Done() = false, want true")
	}
}

func TestSSEParser_BlankLineResetsEventType(t *testing.T) {
	input := "event: typeA\ndata: first\n\ndata: second\n\n"
	p := NewSSEParser(strings.NewReader(input))

	evt, _ := p.Next()
	if evt.Event != "typeA" {
		t.Errorf("first event type = %q, want %q", evt.Event, "typeA")
	}

	evt, _ = p.Next()
	if evt.Event != "" {
		t.Errorf("second event type = %q, want empty (reset by blank line)", evt.Event)
	}
}

func TestSSEParser_EmptyStream(t *testing.T) {
	p := NewSSEParser(strings.NewReader(""))

	_, err := p.Next()
	if err != io.EOF {
		t.Errorf("empty stream: got %v, want io.EOF", err)
	}
}

func TestSSEParser_CommentsIgnored(t *testing.T) {
	input := ": this is a comment\ndata: payload\n\n"
	p := NewSSEParser(strings.NewReader(input))

	evt, err := p.Next()
	if err != nil {
		t.Fatalf("event: %v", err)
	}
	if evt.Data != "payload" {
		t.Errorf("data = %q, want %q", evt.Data, "payload")
	}
}

func TestSSEParser_StreamEndsWithoutTermination(t *testing.T) {
	input := "data: only\n\n"
	p := NewSSEParser(strings.NewReader(input))

	evt, err := p.Next()
	if err != nil {
		t.Fatalf("event: %v", err)
	}
	if evt.Data != "only" {
		t.Fatalf("data = %q, want %q", evt.Data, "only")
	}

	_, err = p.Next()
	if err != io.EOF {
		t.Errorf("stream end: got %v, want io.EOF", err)
	}
	// Done() should be false — no explicit termination signal was seen.
	if p.Done() {
		t.Error("Done() = true, want false (no termination signal)")
	}
}
