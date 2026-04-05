package streaming

import (
	"encoding/json"

	"github.com/kw12121212/spec-coding-sdk/internal/llm"
)

// pendingTC holds partially accumulated tool call data.
type pendingTC struct {
	ID      string
	Name    string
	Partial string // accumulated JSON fragments
}

// ToolCallAccumulator buffers incremental tool-call data across streaming
// chunks and yields complete ToolCalls once all fragments for a given call
// have been received.
type ToolCallAccumulator struct {
	pendings map[int]*pendingTC // keyed by tool call index
}

// NewToolCallAccumulator creates a ready-to-use accumulator.
func NewToolCallAccumulator() *ToolCallAccumulator {
	return &ToolCallAccumulator{
		pendings: make(map[int]*pendingTC),
	}
}

// FeedPartial appends a partial JSON fragment for the tool call at the given
// index. It returns a list of tool calls that are now complete.
//
//   - id and name are captured from the first non-empty values seen for that
//     index (OpenAI sends them in the first delta chunk).
//   - final reports whether this index will receive no more data (i.e. the
//     provider has signalled the end of this tool call's input).
func (a *ToolCallAccumulator) FeedPartial(index int, id, name, partialJSON string, final bool) []llm.ToolCall {
	p, ok := a.pendings[index]
	if !ok {
		p = &pendingTC{}
		a.pendings[index] = p
	}
	if id != "" {
		p.ID = id
	}
	if name != "" {
		p.Name = name
	}
	p.Partial += partialJSON

	if !final {
		return nil
	}

	// Tool call is complete — build and remove from pending.
	tc := llm.ToolCall{
		ID:    p.ID,
		Name:  p.Name,
		Input: json.RawMessage(p.Partial),
	}
	delete(a.pendings, index)
	return []llm.ToolCall{tc}
}

// FeedChunk is a convenience for OpenAI-style delta chunks where tool calls
// arrive as a list indexed by position.  It feeds each entry and returns all
// newly completed tool calls.
//
// Each element of toolCalls must have ID, Name, and Input populated.
// Input may contain partial JSON; set the corresponding final flag to true
// when the provider signals the last fragment for that index.
func (a *ToolCallAccumulator) FeedChunk(toolCalls []llm.ToolCall, finals []bool) []llm.ToolCall {
	var completed []llm.ToolCall
	for i, tc := range toolCalls {
		isFinal := false
		if i < len(finals) {
			isFinal = finals[i]
		}
		done := a.FeedPartial(i, tc.ID, tc.Name, string(tc.Input), isFinal)
		completed = append(completed, done...)
	}
	return completed
}

// Flush returns all pending tool calls (including incomplete ones) and clears
// the accumulator state. This should be called at the end of a stream for
// cleanup or error handling.
func (a *ToolCallAccumulator) Flush() []llm.ToolCall {
	if len(a.pendings) == 0 {
		return nil
	}
	var results []llm.ToolCall
	for idx, p := range a.pendings {
		results = append(results, llm.ToolCall{
			ID:    p.ID,
			Name:  p.Name,
			Input: json.RawMessage(p.Partial),
		})
		delete(a.pendings, idx)
	}
	return results
}
