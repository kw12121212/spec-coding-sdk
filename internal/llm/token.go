package llm

import (
	"encoding/json"
	"errors"
)

// charsPerToken is the heuristic ratio used to estimate token count from character count.
const charsPerToken = 4

// EstimateTokens returns a heuristic estimate of the token count for the given messages.
// It uses a ~4 characters per token ratio and includes message content and tool call JSON.
// Returns 0 for an empty message list.
func EstimateTokens(messages []Message) int {
	totalChars := 0
	for i := range messages {
		totalChars += len(messages[i].Content)
		if len(messages[i].ToolCalls) > 0 {
			data, err := json.Marshal(messages[i].ToolCalls)
			if err == nil {
				totalChars += len(data)
			}
		}
	}
	if totalChars == 0 {
		return 0
	}
	return (totalChars + charsPerToken - 1) / charsPerToken
}

// ModelContextWindow describes the context window size for a specific model.
type ModelContextWindow struct {
	ModelID     string
	TotalTokens int
}

// modelContextWindows maps known model IDs to their context window sizes.
var modelContextWindows = map[string]int{
	"gpt-4o":            128_000,
	"gpt-4-turbo":       128_000,
	"gpt-3.5-turbo":     16_385,
	"claude-sonnet-4-6": 200_000,
	"claude-opus-4-6":   200_000,
	"claude-haiku-4-5":  200_000,
}

// ContextWindow returns the context window size for the given model ID.
// Returns (size, true) for known models, or (0, false) for unknown models.
func ContextWindow(modelID string) (int, bool) {
	size, ok := modelContextWindows[modelID]
	return size, ok
}

// ContextChecker provides methods to check if messages fit within a model's context window.
type ContextChecker struct{}

// Fits checks whether the given messages fit within the model's context window
// after reserving tokens for the response.
// Returns false if the model is unknown or the messages exceed the window.
func (ContextChecker) Fits(messages []Message, model string, reserved int) bool {
	window, ok := ContextWindow(model)
	if !ok {
		return false
	}
	return EstimateTokens(messages)+reserved <= window
}

// ErrUnknownModel is returned when the model ID is not found in the registry.
var ErrUnknownModel = errors.New("unknown model")

// Remaining returns the number of tokens remaining in the model's context window
// after accounting for the messages and reserved tokens.
// Returns an error if the model is unknown.
func (ContextChecker) Remaining(messages []Message, model string, reserved int) (int, error) {
	window, ok := ContextWindow(model)
	if !ok {
		return 0, ErrUnknownModel
	}
	return window - EstimateTokens(messages) - reserved, nil
}
