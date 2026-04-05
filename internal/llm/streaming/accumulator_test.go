package streaming

import (
	"encoding/json"
	"testing"

	"github.com/kw12121212/spec-coding-sdk/internal/llm"
)

func TestAccumulator_SingleChunkComplete(t *testing.T) {
	acc := NewToolCallAccumulator()

	completed := acc.FeedPartial(0, "tc-1", "get_weather", `{"city":"Paris"}`, true)
	if len(completed) != 1 {
		t.Fatalf("completed = %d, want 1", len(completed))
	}
	tc := completed[0]
	if tc.ID != "tc-1" {
		t.Errorf("ID = %q, want %q", tc.ID, "tc-1")
	}
	if tc.Name != "get_weather" {
		t.Errorf("Name = %q, want %q", tc.Name, "get_weather")
	}
	if string(tc.Input) != `{"city":"Paris"}` {
		t.Errorf("Input = %q, want %q", string(tc.Input), `{"city":"Paris"}`)
	}
}

func TestAccumulator_MultiChunkAssembly(t *testing.T) {
	acc := NewToolCallAccumulator()

	// First chunk: ID + Name, partial arguments.
	completed := acc.FeedPartial(0, "tc-1", "search", `{"qu`, false)
	if len(completed) != 0 {
		t.Errorf("first chunk: completed = %d, want 0", len(completed))
	}

	// Second chunk: more arguments.
	completed = acc.FeedPartial(0, "", "", `ery":"`, false)
	if len(completed) != 0 {
		t.Errorf("second chunk: completed = %d, want 0", len(completed))
	}

	// Third (final) chunk: rest of arguments.
	completed = acc.FeedPartial(0, "", "", `test"}`, true)
	if len(completed) != 1 {
		t.Fatalf("final chunk: completed = %d, want 1", len(completed))
	}

	tc := completed[0]
	if tc.ID != "tc-1" {
		t.Errorf("ID = %q, want %q", tc.ID, "tc-1")
	}
	if tc.Name != "search" {
		t.Errorf("Name = %q, want %q", tc.Name, "search")
	}
	want := `{"query":"test"}`
	if string(tc.Input) != want {
		t.Errorf("Input = %q, want %q", string(tc.Input), want)
	}
}

func TestAccumulator_MultipleConcurrentToolCalls(t *testing.T) {
	acc := NewToolCallAccumulator()

	// Interleaved chunks for two tool calls.
	acc.FeedPartial(0, "tc-a", "read_file", `{"pa`, false)
	acc.FeedPartial(1, "tc-b", "write_file", `{"co`, false)
	acc.FeedPartial(0, "", "", `th":"/tmp"}`, true)

	completed := acc.FeedPartial(1, "", "", `ntent":"x"}`, true)
	if len(completed) != 1 {
		t.Fatalf("completed = %d, want 1 (only tc-b)", len(completed))
	}
	if completed[0].ID != "tc-b" {
		t.Errorf("ID = %q, want %q", completed[0].ID, "tc-b")
	}

	// tc-a was already completed; flushing should yield nothing.
	remaining := acc.Flush()
	if len(remaining) != 0 {
		t.Errorf("flush = %d, want 0", len(remaining))
	}
}

func TestAccumulator_FlushIncompleteData(t *testing.T) {
	acc := NewToolCallAccumulator()

	acc.FeedPartial(0, "tc-1", "tool_a", `{"partial`, false)
	acc.FeedPartial(1, "tc-2", "tool_b", `{"incomplete`, false)

	remaining := acc.Flush()
	if len(remaining) != 2 {
		t.Fatalf("flush = %d items, want 2", len(remaining))
	}

	// Verify content is preserved.
	found := map[string]string{}
	for _, tc := range remaining {
		found[tc.ID] = string(tc.Input)
	}
	if found["tc-1"] != `{"partial` {
		t.Errorf("tc-1 input = %q", found["tc-1"])
	}
	if found["tc-2"] != `{"incomplete` {
		t.Errorf("tc-2 input = %q", found["tc-2"])
	}

	// After flush, accumulator should be clean.
	remaining2 := acc.Flush()
	if len(remaining2) != 0 {
		t.Errorf("second flush = %d, want 0", len(remaining2))
	}
}

func TestAccumulator_FeedChunk(t *testing.T) {
	acc := NewToolCallAccumulator()

	// Simulate OpenAI-style two-delta tool call accumulation.
	delta1 := []llm.ToolCall{
		{ID: "tc-1", Name: "calc", Input: json.RawMessage(`{"ex`)},
	}
	finals1 := []bool{false}
	completed := acc.FeedChunk(delta1, finals1)
	if len(completed) != 0 {
		t.Errorf("delta1: completed = %d, want 0", len(completed))
	}

	delta2 := []llm.ToolCall{
		{ID: "", Name: "", Input: json.RawMessage(`pr":1}`)},
	}
	finals2 := []bool{true}
	completed = acc.FeedChunk(delta2, finals2)
	if len(completed) != 1 {
		t.Fatalf("delta2: completed = %d, want 1", len(completed))
	}
	if string(completed[0].Input) != `{"expr":1}` {
		t.Errorf("assembled input = %q", string(completed[0].Input))
	}
}

func TestAccumulator_FlushEmpty(t *testing.T) {
	acc := NewToolCallAccumulator()
	remaining := acc.Flush()
	if remaining != nil {
		t.Errorf("empty flush = %v, want nil", remaining)
	}
}
