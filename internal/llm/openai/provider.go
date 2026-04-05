// Package openai implements the LLM Provider interface for OpenAI-compatible APIs.
package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/kw12121212/spec-coding-sdk/internal/llm"
	"github.com/kw12121212/spec-coding-sdk/internal/llm/streaming"
)

// Compile-time check that OpenAIProvider satisfies llm.Provider.
var _ llm.Provider = (*OpenAIProvider)(nil)

// Option configures an OpenAIProvider.
type Option func(*OpenAIProvider)

// WithHTTPClient sets a custom HTTP client on the provider.
func WithHTTPClient(client *http.Client) Option {
	return func(p *OpenAIProvider) {
		p.client = client
	}
}

// OpenAIProvider implements llm.Provider for OpenAI-compatible Chat Completions APIs.
type OpenAIProvider struct {
	config llm.ProviderConfig
	client *http.Client
}

// NewProvider creates a new OpenAI-compatible provider with the given config and options.
func NewProvider(config llm.ProviderConfig, opts ...Option) *OpenAIProvider {
	p := &OpenAIProvider{
		config: config,
		client: http.DefaultClient,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// --- OpenAI wire types ---

type chatRequest struct {
	Model       string           `json:"model"`
	Messages    []chatMessage    `json:"messages"`
	Temperature *float64         `json:"temperature,omitempty"`
	MaxTokens   *int             `json:"max_tokens,omitempty"`
	Stream      bool             `json:"stream"`
}

type chatMessage struct {
	Role       string           `json:"role"`
	Content    string           `json:"content"`
	ToolCalls  []chatToolCall   `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
}

type chatToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function chatFunction     `json:"function"`
}

type chatFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type chatResponse struct {
	Choices []chatChoice `json:"choices"`
	Usage   chatUsage    `json:"usage"`
}

type chatChoice struct {
	Index        int          `json:"index"`
	Message      chatMessage  `json:"message"`
	FinishReason string       `json:"finish_reason"`
	Delta        *chatDelta   `json:"delta,omitempty"`
}

type chatDelta struct {
	Role      string           `json:"role,omitempty"`
	Content   string           `json:"content,omitempty"`
	ToolCalls []chatToolCall   `json:"tool_calls,omitempty"`
}

type chatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}

type errorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

// --- Format conversion ---

func toChatRequest(req llm.Request, config llm.ProviderConfig, stream bool) chatRequest {
	model := req.Model
	if model == "" {
		model = config.Model
	}

	messages := make([]chatMessage, len(req.Messages))
	for i, m := range req.Messages {
		cm := chatMessage{
			Role:    string(m.Role),
			Content: m.Content,
		}
		if m.Role == llm.RoleTool {
			cm.ToolCallID = m.ToolCallID
		}
		if m.Role == llm.RoleAssistant && len(m.ToolCalls) > 0 {
			cm.ToolCalls = make([]chatToolCall, len(m.ToolCalls))
			for j, tc := range m.ToolCalls {
				cm.ToolCalls[j] = chatToolCall{
					ID:   tc.ID,
					Type: "function",
					Function: chatFunction{
						Name:      tc.Name,
						Arguments: string(tc.Input),
					},
				}
			}
		}
		messages[i] = cm
	}

	cr := chatRequest{
		Model:    model,
		Messages: messages,
		Stream:   stream,
	}
	if req.Temperature != 0 {
		cr.Temperature = &req.Temperature
	}
	if req.MaxTokens != 0 {
		cr.MaxTokens = &req.MaxTokens
	}
	return cr
}

func mapFinishReason(reason string) string {
	switch reason {
	case "stop":
		return "stop"
	case "tool_calls":
		return "tool_use"
	case "length":
		return "max_tokens"
	default:
		return reason
	}
}

func toResponse(cr chatResponse) llm.Response {
	if len(cr.Choices) == 0 {
		return llm.Response{
			Usage: llm.Usage{
				PromptTokens:     cr.Usage.PromptTokens,
				CompletionTokens: cr.Usage.CompletionTokens,
			},
		}
	}

	choice := cr.Choices[0]
	resp := llm.Response{
		Content:    choice.Message.Content,
		StopReason: mapFinishReason(choice.FinishReason),
		Usage: llm.Usage{
			PromptTokens:     cr.Usage.PromptTokens,
			CompletionTokens: cr.Usage.CompletionTokens,
		},
	}

	for _, tc := range choice.Message.ToolCalls {
		resp.ToolCalls = append(resp.ToolCalls, llm.ToolCall{
			ID:    tc.ID,
			Name:  tc.Function.Name,
			Input: json.RawMessage(tc.Function.Arguments),
		})
	}
	return resp
}

// --- HTTP helpers ---

func (p *OpenAIProvider) endpoint() string {
	base := strings.TrimRight(p.config.BaseURL, "/")
	return base + "/chat/completions"
}

func (p *OpenAIProvider) newHTTPRequest(ctx context.Context, body chatRequest) (*http.Request, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint(), strings.NewReader(string(data)))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	return req, nil
}

func parseAPIError(statusCode int, body io.Reader) error {
	data, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("API error (status %d): failed to read body: %w", statusCode, err)
	}
	var errResp errorResponse
	if err := json.Unmarshal(data, &errResp); err != nil {
		return fmt.Errorf("API error (status %d): %s", statusCode, strings.TrimSpace(string(data)))
	}
	return fmt.Errorf("API error (status %d): %s", statusCode, errResp.Error.Message)
}

// --- Provider interface ---

// Complete sends a synchronous Chat Completions request and returns the full response.
func (p *OpenAIProvider) Complete(ctx context.Context, req llm.Request) (llm.Response, error) {
	body := toChatRequest(req, p.config, false)
	httpReq, err := p.newHTTPRequest(ctx, body)
	if err != nil {
		return llm.Response{}, err
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return llm.Response{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return llm.Response{}, parseAPIError(resp.StatusCode, resp.Body)
	}

	var chatResp chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return llm.Response{}, fmt.Errorf("decode response: %w", err)
	}
	return toResponse(chatResp), nil
}

// Stream sends a streaming Chat Completions request and delivers chunks via the callback.
func (p *OpenAIProvider) Stream(ctx context.Context, req llm.Request, callback llm.StreamCallback) error {
	body := toChatRequest(req, p.config, true)
	httpReq, err := p.newHTTPRequest(ctx, body)
	if err != nil {
		return err
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return parseAPIError(resp.StatusCode, resp.Body)
	}

	parser := streaming.NewSSEParser(resp.Body)
	acc := streaming.NewToolCallAccumulator()

	for {
		evt, err := parser.Next()
		if err != nil {
			// Stream ended — flush any remaining partial tool calls.
			if remaining := acc.Flush(); len(remaining) > 0 {
				_ = callback(llm.StreamChunk{ToolCalls: remaining})
			}
			if err == io.EOF {
				return nil
			}
			return err
		}

		var chunk chatResponse
		if err := json.Unmarshal([]byte(evt.Data), &chunk); err != nil {
			return fmt.Errorf("parse SSE chunk: %w", err)
		}

		if len(chunk.Choices) == 0 {
			continue
		}

		choice := chunk.Choices[0]
		if choice.Delta == nil {
			continue
		}

		sChunk := llm.StreamChunk{
			Content: choice.Delta.Content,
		}

		// Accumulate incremental tool calls.
		if len(choice.Delta.ToolCalls) > 0 {
			deltaCalls := make([]llm.ToolCall, len(choice.Delta.ToolCalls))
			finals := make([]bool, len(choice.Delta.ToolCalls))
			for i, tc := range choice.Delta.ToolCalls {
				deltaCalls[i] = llm.ToolCall{
					ID:    tc.ID,
					Name:  tc.Function.Name,
					Input: json.RawMessage(tc.Function.Arguments),
				}
				// OpenAI sends finish_reason="tool_calls" on the final chunk
				// that carries the last arguments fragment.
				finals[i] = choice.FinishReason == "tool_calls"
			}
			completed := acc.FeedChunk(deltaCalls, finals)
			sChunk.ToolCalls = completed
		}

		if chunk.Usage.PromptTokens > 0 || chunk.Usage.CompletionTokens > 0 {
			sChunk.Usage = llm.Usage{
				PromptTokens:     chunk.Usage.PromptTokens,
				CompletionTokens: chunk.Usage.CompletionTokens,
			}
		}

		if sChunk.Content != "" || len(sChunk.ToolCalls) > 0 || sChunk.Usage.PromptTokens != 0 || sChunk.Usage.CompletionTokens != 0 {
			if err := callback(sChunk); err != nil {
				return err
			}
		}
	}
}
