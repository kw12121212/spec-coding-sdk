// Package streaming provides shared SSE parsing and tool-call accumulation
// utilities used by LLM provider implementations.
package streaming

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// SSEEvent represents a single Server-Sent Event.
type SSEEvent struct {
	Event string // value from "event:" line (empty if absent)
	Data  string // value from "data:" line
}

// SSEParser reads SSE frames from an io.Reader.
type SSEParser struct {
	scanner   *bufio.Scanner
	done      bool
	eventType string
}

// NewSSEParser creates a parser that reads from r.
func NewSSEParser(r io.Reader) *SSEParser {
	s := bufio.NewScanner(r)
	s.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	return &SSEParser{scanner: s}
}

// Next returns the next SSE event.
// It returns io.EOF when the stream ends normally (data: [DONE] or
// event: message_stop). Any other error indicates a parse or read failure.
func (p *SSEParser) Next() (SSEEvent, error) {
	for p.scanner.Scan() {
		line := p.scanner.Text()

		// Blank line resets event type (SSE spec).
		if line == "" {
			p.eventType = ""
			continue
		}

		// event: <type>
		if evt, ok := strings.CutPrefix(line, "event: "); ok {
			p.eventType = evt
			// Claude termination signal.
			if evt == "message_stop" {
				p.done = true
				return SSEEvent{}, io.EOF
			}
			continue
		}

		// data: <payload>
		data, ok := strings.CutPrefix(line, "data: ")
		if !ok {
			// Unknown line format — skip comments (lines starting with ':') and
			// other SSE fields (id:, retry:) that we don't consume.
			continue
		}

		// OpenAI termination signal.
		if data == "[DONE]" {
			p.done = true
			return SSEEvent{}, io.EOF
		}

		evt := SSEEvent{
			Event: p.eventType,
			Data:  data,
		}
		// Reset event type after data line (standard SSE semantics).
		p.eventType = ""
		return evt, nil
	}

	if err := p.scanner.Err(); err != nil {
		return SSEEvent{}, fmt.Errorf("SSE read error: %w", err)
	}
	// Stream ended without a termination signal.
	return SSEEvent{}, io.EOF
}

// Done reports whether the parser has seen a termination signal.
func (p *SSEParser) Done() bool { return p.done }
