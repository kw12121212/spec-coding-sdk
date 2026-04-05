// Package claude implements the LLM Provider interface for the Anthropic Claude Messages API.
package claude

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/kw12121212/spec-coding-sdk/internal/llm"
)

// Compile-time check that ClaudeProvider satisfies llm.Provider.
var _ llm.Provider = (*ClaudeProvider)(nil)

// Option configures a ClaudeProvider.
type Option func(*ClaudeProvider)

// WithHTTPClient sets a custom HTTP client on the provider.
func WithHTTPClient(client *http.Client) Option {
	return func(p *ClaudeProvider) {
		p.client = client
	}
}

// ClaudeProvider implements llm.Provider for the Anthropic Claude Messages API.
type ClaudeProvider struct {
	config llm.ProviderConfig
	client *http.Client
}

// NewProvider creates a new Claude provider with the given config and options.
func NewProvider(config llm.ProviderConfig, opts ...Option) *ClaudeProvider {
	p := &ClaudeProvider{
		config: config,
		client: http.DefaultClient,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// --- Claude wire types ---

type messagesRequest struct {
	Model       string              `json:"model"`
	System      string              `json:"system,omitempty"`
	Messages    []claudeMessage     `json:"messages"`
	MaxTokens   int                 `json:"max_tokens"`
	Temperature *float64            `json:"temperature,omitempty"`
	Stream      bool                `json:"stream"`
}

type claudeMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

type textContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type toolUseContentBlock struct {
	Type  string          `json:"type"`
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

type toolResultContentBlock struct {
	Type      string `json:"type"`
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
}

type messagesResponse struct {
	ID         string             `json:"id"`
	Type       string             `json:"type"`
	Role       string             `json:"role"`
	Content    []contentBlock     `json:"content"`
	Model      string             `json:"model"`
	StopReason string             `json:"stop_reason"`
	Usage      responseUsage      `json:"usage"`
}

type contentBlock struct {
	Type  string          `json:"type"`
	Text  string          `json:"text,omitempty"`
	ID    string          `json:"id,omitempty"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`
}

type responseUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type errorResponse struct {
	Type  string `json:"type"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// --- SSE event types ---

type sseContentBlockDelta struct {
	Type  string      `json:"type"`
	Index int         `json:"index"`
	Delta sseDelta    `json:"delta"`
}

type sseDelta struct {
	Type           string          `json:"type"`
	Text           string          `json:"text,omitempty"`
	PartialJSON    string          `json:"partial_json,omitempty"`
	StopReason     string          `json:"stop_reason,omitempty"`
	InputTokens    int             `json:"input_tokens,omitempty"`
	OutputTokens   int             `json:"output_tokens,omitempty"`
}

type sseMessageDelta struct {
	Type       string      `json:"type"`
	Delta      sseDelta    `json:"delta"`
	Usage      responseUsage `json:"usage"`
}

// --- Format conversion ---

func toMessagesRequest(req llm.Request, config llm.ProviderConfig, stream bool) messagesRequest {
	model := req.Model
	if model == "" {
		model = config.Model
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	var system string
	var messages []claudeMessage

	for _, m := range req.Messages {
		if string(m.Role) == "system" {
			system = m.Content
			continue
		}

		cm := claudeMessage{Role: string(m.Role)}

		switch m.Role {
		case llm.RoleTool:
			// Tool result: mapped as user message with tool_result content block
			cm.Role = "user"
			block := toolResultContentBlock{
				Type:      "tool_result",
				ToolUseID: m.ToolCallID,
				Content:   m.Content,
			}
			data, _ := json.Marshal([]toolResultContentBlock{block})
			cm.Content = data
		case llm.RoleUser:
			if m.ToolCallID != "" {
				// Tool result: user message with tool_result content block
				block := toolResultContentBlock{
					Type:      "tool_result",
					ToolUseID: m.ToolCallID,
					Content:   m.Content,
				}
				data, _ := json.Marshal([]toolResultContentBlock{block})
				cm.Content = data
			} else {
				block := textContentBlock{Type: "text", Text: m.Content}
				data, _ := json.Marshal([]textContentBlock{block})
				cm.Content = data
			}
		case llm.RoleAssistant:
			if len(m.ToolCalls) > 0 {
				var blocks []json.RawMessage
				if m.Content != "" {
					tb := textContentBlock{Type: "text", Text: m.Content}
					data, _ := json.Marshal(tb)
					blocks = append(blocks, data)
				}
				for _, tc := range m.ToolCalls {
					tu := toolUseContentBlock{
						Type:  "tool_use",
						ID:    tc.ID,
						Name:  tc.Name,
						Input: tc.Input,
					}
					data, _ := json.Marshal(tu)
					blocks = append(blocks, data)
				}
				data, _ := json.Marshal(blocks)
				cm.Content = data
			} else {
				block := textContentBlock{Type: "text", Text: m.Content}
				data, _ := json.Marshal([]textContentBlock{block})
				cm.Content = data
			}
		default:
			block := textContentBlock{Type: "text", Text: m.Content}
			data, _ := json.Marshal([]textContentBlock{block})
			cm.Content = data
		}

		messages = append(messages, cm)
	}

	mr := messagesRequest{
		Model:     model,
		System:    system,
		Messages:  messages,
		MaxTokens: maxTokens,
		Stream:    stream,
	}
	if req.Temperature != 0 {
		mr.Temperature = &req.Temperature
	}
	return mr
}

func mapStopReason(reason string) string {
	switch reason {
	case "end_turn":
		return "stop"
	case "tool_use":
		return "tool_use"
	case "max_tokens":
		return "max_tokens"
	default:
		return reason
	}
}

func toResponse(mr messagesResponse) llm.Response {
	var content strings.Builder
	var toolCalls []llm.ToolCall

	for _, block := range mr.Content {
		switch block.Type {
		case "text":
			content.WriteString(block.Text)
		case "tool_use":
			toolCalls = append(toolCalls, llm.ToolCall{
				ID:    block.ID,
				Name:  block.Name,
				Input: block.Input,
			})
		}
	}

	return llm.Response{
		Content:    content.String(),
		ToolCalls:  toolCalls,
		StopReason: mapStopReason(mr.StopReason),
		Usage: llm.Usage{
			PromptTokens:     mr.Usage.InputTokens,
			CompletionTokens: mr.Usage.OutputTokens,
		},
	}
}

// --- HTTP helpers ---

func (p *ClaudeProvider) endpoint() string {
	base := strings.TrimRight(p.config.BaseURL, "/")
	return base + "/v1/messages"
}

func (p *ClaudeProvider) newHTTPRequest(ctx context.Context, body messagesRequest) (*http.Request, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint(), strings.NewReader(string(data)))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.config.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")
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

// Complete sends a synchronous Messages API request and returns the full response.
func (p *ClaudeProvider) Complete(ctx context.Context, req llm.Request) (llm.Response, error) {
	body := toMessagesRequest(req, p.config, false)
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

	var msgResp messagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&msgResp); err != nil {
		return llm.Response{}, fmt.Errorf("decode response: %w", err)
	}
	return toResponse(msgResp), nil
}

// Stream sends a streaming Messages API request and delivers chunks via the callback.
func (p *ClaudeProvider) Stream(ctx context.Context, req llm.Request, callback llm.StreamCallback) error {
	body := toMessagesRequest(req, p.config, true)
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

	scanner := bufio.NewScanner(resp.Body)
	var eventType string

	for scanner.Scan() {
		line := scanner.Text()

		if evt, ok := strings.CutPrefix(line, "event: "); ok {
			eventType = evt
			continue
		}

		data, ok := strings.CutPrefix(line, "data: ")
		if !ok {
			continue
		}

		switch eventType {
		case "content_block_delta":
			var event sseContentBlockDelta
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				return fmt.Errorf("parse content_block_delta: %w", err)
			}

			chunk := llm.StreamChunk{}
			switch event.Delta.Type {
			case "text_delta":
				chunk.Content = event.Delta.Text
			case "input_json_delta":
				chunk.ToolCalls = []llm.ToolCall{
					{Input: json.RawMessage(event.Delta.PartialJSON)},
				}
			}

			if err := callback(chunk); err != nil {
				return err
			}

		case "message_delta":
			var event sseMessageDelta
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				return fmt.Errorf("parse message_delta: %w", err)
			}

			chunk := llm.StreamChunk{
				Usage: llm.Usage{
					PromptTokens:     event.Usage.InputTokens,
					CompletionTokens: event.Usage.OutputTokens,
				},
			}

			if err := callback(chunk); err != nil {
				return err
			}

		case "message_stop":
			return nil
		}
	}
	return scanner.Err()
}
