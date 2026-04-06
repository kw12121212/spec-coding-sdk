package claude

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kw12121212/spec-coding-sdk/internal/llm"
)

func newTestServer(handler http.HandlerFunc) (*ClaudeProvider, *httptest.Server) {
	ts := httptest.NewServer(handler)
	p := NewProvider(
		llm.ProviderConfig{
			BaseURL: ts.URL,
			APIKey:  "test-key",
			Model:   "claude-sonnet-4-6",
		},
		WithHTTPClient(ts.Client()),
	)
	return p, ts
}

func TestComplete_NormalTextResponse(t *testing.T) {
	p, ts := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("x-api-key") != "test-key" {
			t.Errorf("expected x-api-key header")
		}
		if r.Header.Get("anthropic-version") != "2023-06-01" {
			t.Errorf("expected anthropic-version header")
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected application/json content type")
		}

		var req messagesRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Stream {
			t.Error("expected stream=false for Complete")
		}
		if req.Model != "claude-sonnet-4-6" {
			t.Errorf("expected model claude-sonnet-4-6, got %s", req.Model)
		}
		if len(req.Messages) != 1 || req.Messages[0].Role != "user" {
			t.Errorf("expected 1 user message, got %+v", req.Messages)
		}

		resp := messagesResponse{
			Content: []contentBlock{
				{Type: "text", Text: "Hello!"},
			},
			StopReason: "end_turn",
			Usage:      responseUsage{InputTokens: 10, OutputTokens: 5},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer ts.Close()

	resp, err := p.Complete(context.Background(), llm.Request{
		Model:    "claude-sonnet-4-6",
		Messages: []llm.Message{{Role: llm.RoleUser, Content: "Hi"}},
	})
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if resp.Content != "Hello!" {
		t.Errorf("content: got %q, want %q", resp.Content, "Hello!")
	}
	if resp.StopReason != "stop" {
		t.Errorf("stop_reason: got %q, want %q", resp.StopReason, "stop")
	}
	if resp.Usage.PromptTokens != 10 || resp.Usage.CompletionTokens != 5 {
		t.Errorf("usage: got %+v", resp.Usage)
	}
	if len(resp.ToolCalls) != 0 {
		t.Errorf("expected no tool calls, got %d", len(resp.ToolCalls))
	}
}

func TestComplete_ToolUseResponse(t *testing.T) {
	p, ts := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := messagesResponse{
			Content: []contentBlock{
				{Type: "text", Text: "Let me run that."},
				{
					Type:  "tool_use",
					ID:    "toolu_abc",
					Name:  "bash",
					Input: json.RawMessage(`{"command":"ls"}`),
				},
			},
			StopReason: "tool_use",
			Usage:      responseUsage{InputTokens: 20, OutputTokens: 10},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer ts.Close()

	resp, err := p.Complete(context.Background(), llm.Request{
		Messages: []llm.Message{{Role: llm.RoleUser, Content: "run ls"}},
	})
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if resp.StopReason != "tool_use" {
		t.Errorf("stop_reason: got %q, want %q", resp.StopReason, "tool_use")
	}
	if len(resp.ToolCalls) != 1 {
		t.Fatalf("tool_calls: got %d, want 1", len(resp.ToolCalls))
	}
	tc := resp.ToolCalls[0]
	if tc.ID != "toolu_abc" || tc.Name != "bash" {
		t.Errorf("tool_call: got id=%q name=%q", tc.ID, tc.Name)
	}
	if string(tc.Input) != `{"command":"ls"}` {
		t.Errorf("tool_call input: got %s", tc.Input)
	}
	if resp.Content != "Let me run that." {
		t.Errorf("content: got %q, want %q", resp.Content, "Let me run that.")
	}
}

func TestStream_TextChunks(t *testing.T) {
	p, ts := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req messagesRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if !req.Stream {
			t.Error("expected stream=true for Stream")
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		events := []string{
			"event: message_start\ndata: {\"type\":\"message_start\"}\n\n",
			"event: content_block_start\ndata: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"text\",\"text\":\"\"}}\n\n",
			"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"Hello\"}}\n\n",
			"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\" world\"}}\n\n",
			"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"!\"}}\n\n",
			"event: content_block_stop\ndata: {\"type\":\"content_block_stop\",\"index\":0}\n\n",
			"event: message_delta\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"end_turn\"},\"usage\":{\"output_tokens\":5}}\n\n",
			"event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n",
		}
		for _, evt := range events {
			if _, err := fmt.Fprint(w, evt); err != nil {
				t.Fatalf("write event: %v", err)
			}
		}
	}))
	defer ts.Close()

	var contents []string
	var gotUsage bool
	err := p.Stream(context.Background(), llm.Request{
		Messages: []llm.Message{{Role: llm.RoleUser, Content: "Hi"}},
	}, func(chunk llm.StreamChunk) error {
		contents = append(contents, chunk.Content)
		if chunk.Usage.CompletionTokens > 0 {
			gotUsage = true
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}
	got := strings.Join(contents, "")
	if got != "Hello world!" {
		t.Errorf("stream content: got %q, want %q", got, "Hello world!")
	}
	if !gotUsage {
		t.Error("expected usage in message_delta chunk")
	}
}

func TestStream_ToolCallChunks(t *testing.T) {
	p, ts := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		events := []string{
			"event: content_block_start\ndata: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"text\",\"text\":\"\"}}\n\n",
			"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"Using tool\"}}\n\n",
			"event: content_block_stop\ndata: {\"type\":\"content_block_stop\",\"index\":0}\n\n",
			"event: content_block_start\ndata: {\"type\":\"content_block_start\",\"index\":1,\"content_block\":{\"type\":\"tool_use\",\"id\":\"tu-1\",\"name\":\"bash\"}}\n\n",
			"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":1,\"delta\":{\"type\":\"input_json_delta\",\"partial_json\":\"{\\\"command\\\":\\\"ls\\\"}\"}}\n\n",
			"event: content_block_stop\ndata: {\"type\":\"content_block_stop\",\"index\":1}\n\n",
			"event: message_delta\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"tool_use\"},\"usage\":{\"output_tokens\":10}}\n\n",
			"event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n",
		}
		for _, evt := range events {
			if _, err := fmt.Fprint(w, evt); err != nil {
				t.Fatalf("write event: %v", err)
			}
		}
	}))
	defer ts.Close()

	var contents []string
	var gotToolCall bool
	err := p.Stream(context.Background(), llm.Request{
		Messages: []llm.Message{{Role: llm.RoleUser, Content: "run ls"}},
	}, func(chunk llm.StreamChunk) error {
		contents = append(contents, chunk.Content)
		if len(chunk.ToolCalls) > 0 {
			gotToolCall = true
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}
	if !gotToolCall {
		t.Error("expected tool call chunk from input_json_delta")
	}
	got := strings.Join(contents, "")
	if got != "Using tool" {
		t.Errorf("stream content: got %q, want %q", got, "Using tool")
	}
}

func TestHTTPError(t *testing.T) {
	p, ts := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		if _, err := w.Write([]byte(`{"type":"error","error":{"type":"rate_limit_error","message":"Rate limit exceeded"}}`)); err != nil {
			t.Fatalf("write error response: %v", err)
		}
	}))
	defer ts.Close()

	_, err := p.Complete(context.Background(), llm.Request{
		Messages: []llm.Message{{Role: llm.RoleUser, Content: "Hi"}},
	})
	if err == nil {
		t.Fatal("expected error for 429 response")
	}
	if !strings.Contains(err.Error(), "429") {
		t.Errorf("error should contain status code: %v", err)
	}
	if !strings.Contains(err.Error(), "Rate limit exceeded") {
		t.Errorf("error should contain API message: %v", err)
	}
}

func TestEmptyMessages(t *testing.T) {
	p, ts := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req messagesRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if len(req.Messages) != 0 {
			t.Errorf("expected empty messages, got %d", len(req.Messages))
		}
		resp := messagesResponse{
			Content: []contentBlock{
				{Type: "text", Text: "ok"},
			},
			StopReason: "end_turn",
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer ts.Close()

	resp, err := p.Complete(context.Background(), llm.Request{})
	if err != nil {
		t.Fatalf("Complete with empty messages: %v", err)
	}
	if resp.Content != "ok" {
		t.Errorf("content: got %q, want %q", resp.Content, "ok")
	}
}

func TestModelFallback(t *testing.T) {
	p, ts := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req messagesRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Model != "claude-sonnet-4-6" {
			t.Errorf("expected fallback to config model claude-sonnet-4-6, got %q", req.Model)
		}
		resp := messagesResponse{
			Content: []contentBlock{
				{Type: "text", Text: "ok"},
			},
			StopReason: "end_turn",
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer ts.Close()

	// Request has empty Model — should fall back to ProviderConfig.Model
	_, err := p.Complete(context.Background(), llm.Request{
		Messages: []llm.Message{{Role: llm.RoleUser, Content: "Hi"}},
	})
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
}

func TestMessageFormatConversion(t *testing.T) {
	p, ts := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req messagesRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		// System message extracted to top-level
		if req.System != "You are a helpful assistant." {
			t.Errorf("system: got %q", req.System)
		}

		// Should have 3 messages (system skipped)
		if len(req.Messages) != 3 {
			t.Fatalf("expected 3 messages (system excluded), got %d", len(req.Messages))
		}

		// user message
		if req.Messages[0].Role != "user" {
			t.Errorf("msg[0] role: got %q", req.Messages[0].Role)
		}
		var userBlocks []textContentBlock
		if err := json.Unmarshal(req.Messages[0].Content, &userBlocks); err != nil {
			t.Fatalf("unmarshal user blocks: %v", err)
		}
		if len(userBlocks) != 1 || userBlocks[0].Text != "run ls" {
			t.Errorf("msg[0] content: got %+v", userBlocks)
		}

		// assistant with tool calls
		if req.Messages[1].Role != "assistant" {
			t.Errorf("msg[1] role: got %q", req.Messages[1].Role)
		}
		var asstBlocks []json.RawMessage
		if err := json.Unmarshal(req.Messages[1].Content, &asstBlocks); err != nil {
			t.Fatalf("unmarshal assistant blocks: %v", err)
		}
		// Only tool_use block (Content is empty, no text block emitted)
		if len(asstBlocks) != 1 {
			t.Fatalf("msg[1] blocks: got %d, want 1", len(asstBlocks))
		}
		var tu toolUseContentBlock
		if err := json.Unmarshal(asstBlocks[0], &tu); err != nil {
			t.Fatalf("unmarshal tool_use block: %v", err)
		}
		if tu.Name != "bash" {
			t.Errorf("msg[1] tool_use name: got %q", tu.Name)
		}

		// tool result — mapped as user with tool_result content block
		if req.Messages[2].Role != "user" {
			t.Errorf("msg[2] role: got %q, want user (tool result)", req.Messages[2].Role)
		}
		var toolResultBlocks []toolResultContentBlock
		if err := json.Unmarshal(req.Messages[2].Content, &toolResultBlocks); err != nil {
			t.Fatalf("unmarshal tool result blocks: %v", err)
		}
		if len(toolResultBlocks) != 1 {
			t.Fatalf("msg[2] tool_result blocks: got %d", len(toolResultBlocks))
		}
		if toolResultBlocks[0].ToolUseID != "toolu_1" {
			t.Errorf("msg[2] tool_use_id: got %q", toolResultBlocks[0].ToolUseID)
		}

		resp := messagesResponse{
			Content: []contentBlock{
				{Type: "text", Text: "ok"},
			},
			StopReason: "end_turn",
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer ts.Close()

	_, err := p.Complete(context.Background(), llm.Request{
		Messages: []llm.Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: llm.RoleUser, Content: "run ls"},
			{Role: llm.RoleAssistant, Content: "", ToolCalls: []llm.ToolCall{
				{ID: "toolu_1", Name: "bash", Input: json.RawMessage(`{"command":"ls"}`)},
			}},
			{Role: llm.RoleTool, Content: "file1.go\nfile2.go", ToolCallID: "toolu_1"},
		},
	})
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
}

func TestMaxTokensDefault(t *testing.T) {
	p, ts := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req messagesRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.MaxTokens != 4096 {
			t.Errorf("max_tokens: got %d, want 4096 (default)", req.MaxTokens)
		}
		resp := messagesResponse{
			Content: []contentBlock{
				{Type: "text", Text: "ok"},
			},
			StopReason: "end_turn",
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer ts.Close()

	_, err := p.Complete(context.Background(), llm.Request{
		Messages: []llm.Message{{Role: llm.RoleUser, Content: "Hi"}},
	})
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
}

func TestCallbackErrorStopsStream(t *testing.T) {
	p, ts := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		events := []string{
			"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"Hello\"}}\n\n",
			"event: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\" world\"}}\n\n",
		}
		for _, evt := range events {
			if _, err := fmt.Fprint(w, evt); err != nil {
				t.Fatalf("write event: %v", err)
			}
		}
	}))
	defer ts.Close()

	errStopped := fmt.Errorf("stop streaming")
	err := p.Stream(context.Background(), llm.Request{
		Messages: []llm.Message{{Role: llm.RoleUser, Content: "Hi"}},
	}, func(_ llm.StreamChunk) error {
		return errStopped
	})
	if err != errStopped {
		t.Errorf("expected callback error, got %v", err)
	}
}

func TestStream_HTTPError(t *testing.T) {
	p, ts := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		if _, err := w.Write([]byte(`{"type":"error","error":{"type":"authentication_error","message":"Invalid API key"}}`)); err != nil {
			t.Fatalf("write error response: %v", err)
		}
	}))
	defer ts.Close()

	err := p.Stream(context.Background(), llm.Request{
		Messages: []llm.Message{{Role: llm.RoleUser, Content: "Hi"}},
	}, func(_ llm.StreamChunk) error { return nil })
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
	if !strings.Contains(err.Error(), "Invalid API key") {
		t.Errorf("error should contain API message: %v", err)
	}
}
