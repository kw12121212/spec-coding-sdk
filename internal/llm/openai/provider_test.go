package openai

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

func newTestServer(handler http.HandlerFunc) (*OpenAIProvider, *httptest.Server) {
	ts := httptest.NewServer(handler)
	p := NewProvider(
		llm.ProviderConfig{
			BaseURL: ts.URL,
			APIKey:  "test-key",
			Model:   "gpt-4",
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
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("expected Bearer auth header")
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected application/json content type")
		}

		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Stream {
			t.Error("expected stream=false for Complete")
		}
		if req.Model != "gpt-4" {
			t.Errorf("expected model gpt-4, got %s", req.Model)
		}
		if len(req.Messages) != 1 || req.Messages[0].Role != "user" {
			t.Errorf("expected 1 user message, got %+v", req.Messages)
		}

		resp := chatResponse{
			Choices: []chatChoice{
				{
					Message:      chatMessage{Role: "assistant", Content: "Hello!"},
					FinishReason: "stop",
				},
			},
			Usage: chatUsage{PromptTokens: 10, CompletionTokens: 5},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer ts.Close()

	resp, err := p.Complete(context.Background(), llm.Request{
		Model:    "gpt-4",
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

func TestComplete_ToolCallResponse(t *testing.T) {
	p, ts := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := chatResponse{
			Choices: []chatChoice{
				{
					Message: chatMessage{
						Role:    "assistant",
						Content: "",
						ToolCalls: []chatToolCall{
							{
								ID:   "call_abc",
								Type: "function",
								Function: chatFunction{
									Name:      "bash",
									Arguments: `{"command":"ls"}`,
								},
							},
						},
					},
					FinishReason: "tool_calls",
				},
			},
			Usage: chatUsage{PromptTokens: 20, CompletionTokens: 10},
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
	if tc.ID != "call_abc" || tc.Name != "bash" {
		t.Errorf("tool_call: got id=%q name=%q", tc.ID, tc.Name)
	}
	if string(tc.Input) != `{"command":"ls"}` {
		t.Errorf("tool_call input: got %s", tc.Input)
	}
}

func TestStream_NormalText(t *testing.T) {
	p, ts := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if !req.Stream {
			t.Error("expected stream=true for Stream")
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		chunks := []string{
			`{"choices":[{"delta":{"content":"Hello"}}],"usage":{}}`,
			`{"choices":[{"delta":{"content":" world"}}],"usage":{}}`,
			`{"choices":[{"delta":{"content":"!"}}],"usage":{"prompt_tokens":10,"completion_tokens":5}}`,
		}
		for _, chunk := range chunks {
			if _, err := fmt.Fprintf(w, "data: %s\n\n", chunk); err != nil {
				t.Fatalf("write chunk: %v", err)
			}
		}
		if _, err := fmt.Fprint(w, "data: [DONE]\n\n"); err != nil {
			t.Fatalf("write done event: %v", err)
		}
	}))
	defer ts.Close()

	var contents []string
	var gotUsage bool
	err := p.Stream(context.Background(), llm.Request{
		Messages: []llm.Message{{Role: llm.RoleUser, Content: "Hi"}},
	}, func(chunk llm.StreamChunk) error {
		contents = append(contents, chunk.Content)
		if chunk.Usage.PromptTokens > 0 {
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
		t.Error("expected usage in last chunk")
	}
}

func TestStream_CallbackError(t *testing.T) {
	p, ts := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		if _, err := fmt.Fprintf(w, "data: %s\n\n", `{"choices":[{"delta":{"content":"Hello"}}],"usage":{}}`); err != nil {
			t.Fatalf("write first chunk: %v", err)
		}
		if _, err := fmt.Fprintf(w, "data: %s\n\n", `{"choices":[{"delta":{"content":" world"}}],"usage":{}}`); err != nil {
			t.Fatalf("write second chunk: %v", err)
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

func TestHTTPError(t *testing.T) {
	p, ts := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		if _, err := w.Write([]byte(`{"error":{"message":"Rate limit exceeded","type":"rate_limit_error","code":"429"}}`)); err != nil {
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

func TestHTTPError_NonJSON(t *testing.T) {
	p, ts := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte("internal server error")); err != nil {
			t.Fatalf("write error response: %v", err)
		}
	}))
	defer ts.Close()

	_, err := p.Complete(context.Background(), llm.Request{
		Messages: []llm.Message{{Role: llm.RoleUser, Content: "Hi"}},
	})
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestEmptyMessages(t *testing.T) {
	p, ts := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if len(req.Messages) != 0 {
			t.Errorf("expected empty messages, got %d", len(req.Messages))
		}
		resp := chatResponse{
			Choices: []chatChoice{
				{Message: chatMessage{Role: "assistant", Content: "ok"}, FinishReason: "stop"},
			},
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

func TestBaseURL_CustomProvider(t *testing.T) {
	customHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := chatResponse{
			Choices: []chatChoice{
				{Message: chatMessage{Role: "assistant", Content: "deepseek response"}, FinishReason: "stop"},
			},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	})
	ts := httptest.NewServer(customHandler)
	defer ts.Close()

	p := NewProvider(
		llm.ProviderConfig{
			BaseURL: ts.URL,
			APIKey:  "ds-key",
			Model:   "deepseek-chat",
		},
		WithHTTPClient(ts.Client()),
	)

	resp, err := p.Complete(context.Background(), llm.Request{
		Messages: []llm.Message{{Role: llm.RoleUser, Content: "Hi"}},
	})
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if resp.Content != "deepseek response" {
		t.Errorf("content: got %q, want %q", resp.Content, "deepseek response")
	}
}

func TestModelFallback(t *testing.T) {
	p, ts := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Model != "gpt-4" {
			t.Errorf("expected fallback to config model gpt-4, got %q", req.Model)
		}
		resp := chatResponse{
			Choices: []chatChoice{
				{Message: chatMessage{Role: "assistant", Content: "ok"}, FinishReason: "stop"},
			},
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
		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if len(req.Messages) != 3 {
			t.Fatalf("expected 3 messages, got %d", len(req.Messages))
		}
		// user message
		if req.Messages[0].Role != "user" || req.Messages[0].Content != "run ls" {
			t.Errorf("msg[0]: got role=%q content=%q", req.Messages[0].Role, req.Messages[0].Content)
		}
		// assistant with tool calls
		if req.Messages[1].Role != "assistant" {
			t.Errorf("msg[1] role: got %q", req.Messages[1].Role)
		}
		if len(req.Messages[1].ToolCalls) != 1 || req.Messages[1].ToolCalls[0].Function.Name != "bash" {
			t.Errorf("msg[1] tool_calls: got %+v", req.Messages[1].ToolCalls)
		}
		if req.Messages[1].ToolCalls[0].Function.Arguments != `{"command":"ls"}` {
			t.Errorf("msg[1] arguments: got %q", req.Messages[1].ToolCalls[0].Function.Arguments)
		}
		// tool result
		if req.Messages[2].Role != "tool" {
			t.Errorf("msg[2] role: got %q", req.Messages[2].Role)
		}
		if req.Messages[2].ToolCallID != "call_1" {
			t.Errorf("msg[2] tool_call_id: got %q", req.Messages[2].ToolCallID)
		}

		resp := chatResponse{
			Choices: []chatChoice{
				{Message: chatMessage{Role: "assistant", Content: "ok"}, FinishReason: "stop"},
			},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("encode response: %v", err)
		}
	}))
	defer ts.Close()

	_, err := p.Complete(context.Background(), llm.Request{
		Messages: []llm.Message{
			{Role: llm.RoleUser, Content: "run ls"},
			{Role: llm.RoleAssistant, Content: "", ToolCalls: []llm.ToolCall{
				{ID: "call_1", Name: "bash", Input: json.RawMessage(`{"command":"ls"}`)},
			}},
			{Role: llm.RoleTool, Content: "file1.go\nfile2.go", ToolCallID: "call_1"},
		},
	})
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
}

func TestOptionalFields(t *testing.T) {
	p, ts := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		if req.Temperature != nil {
			t.Errorf("Temperature should be nil when zero, got %v", req.Temperature)
		}
		if req.MaxTokens != nil {
			t.Errorf("MaxTokens should be nil when zero, got %v", req.MaxTokens)
		}

		resp := chatResponse{
			Choices: []chatChoice{
				{Message: chatMessage{Role: "assistant", Content: "ok"}, FinishReason: "stop"},
			},
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

func TestStream_HTTPError(t *testing.T) {
	p, ts := newTestServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		if _, err := w.Write([]byte(`{"error":{"message":"Invalid API key","type":"authentication_error","code":"401"}}`)); err != nil {
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
