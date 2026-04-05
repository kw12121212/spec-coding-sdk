package llm

import (
	"encoding/json"
	"testing"
)

func TestEstimateTokens_EmptyMessages(t *testing.T) {
	got := EstimateTokens(nil)
	if got != 0 {
		t.Errorf("EstimateTokens(nil) = %d, want 0", got)
	}
	got = EstimateTokens([]Message{})
	if got != 0 {
		t.Errorf("EstimateTokens([]) = %d, want 0", got)
	}
}

func TestEstimateTokens_PlainText(t *testing.T) {
	msg := Message{Role: RoleUser, Content: "Hello, world!"}
	got := EstimateTokens([]Message{msg})
	expected := (len(msg.Content) + 3) / 4 // ceiling division
	if got != expected {
		t.Errorf("EstimateTokens(plain text) = %d, want ~%d", got, expected)
	}
	if got <= 0 {
		t.Errorf("expected positive estimate, got %d", got)
	}
}

func TestEstimateTokens_WithToolCalls(t *testing.T) {
	msg := Message{
		Role:    RoleAssistant,
		Content: "I will run a command",
		ToolCalls: []ToolCall{
			{ID: "call_1", Name: "bash", Input: json.RawMessage(`{"command":"ls -la"}`)},
		},
	}
	got := EstimateTokens([]Message{msg})
	if got <= 0 {
		t.Errorf("expected positive estimate, got %d", got)
	}
	// Verify tool call JSON is counted: estimate with tool calls must exceed content-only estimate.
	contentOnly := (len(msg.Content) + 3) / 4
	if got <= contentOnly {
		t.Errorf("estimate with tool calls (%d) should exceed content-only (%d)", got, contentOnly)
	}
}

func TestEstimateTokens_MultipleMessages(t *testing.T) {
	messages := []Message{
		{Role: RoleUser, Content: "Hello"},
		{Role: RoleAssistant, Content: "Hi there!"},
		{Role: RoleUser, Content: "How are you?"},
	}
	got := EstimateTokens(messages)
	totalChars := 0
	for _, m := range messages {
		totalChars += len(m.Content)
	}
	expected := (totalChars + 3) / 4
	if got != expected {
		t.Errorf("EstimateTokens(multi) = %d, want %d", got, expected)
	}
}

func TestContextWindow_KnownModels(t *testing.T) {
	tests := []struct {
		model    string
		expected int
	}{
		{"gpt-4o", 128_000},
		{"gpt-4-turbo", 128_000},
		{"gpt-3.5-turbo", 16_385},
		{"claude-sonnet-4-6", 200_000},
		{"claude-opus-4-6", 200_000},
		{"claude-haiku-4-5", 200_000},
	}
	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			size, ok := ContextWindow(tt.model)
			if !ok {
				t.Fatalf("expected model %q to be known", tt.model)
			}
			if size != tt.expected {
				t.Errorf("ContextWindow(%q) = %d, want %d", tt.model, size, tt.expected)
			}
		})
	}
}

func TestContextWindow_UnknownModel(t *testing.T) {
	size, ok := ContextWindow("nonexistent-model")
	if ok {
		t.Errorf("expected unknown model to return false, got size=%d", size)
	}
	if size != 0 {
		t.Errorf("expected size 0 for unknown model, got %d", size)
	}
}

func TestContextChecker_Fits_WithinWindow(t *testing.T) {
	checker := ContextChecker{}
	// Short message should fit in gpt-4o's 128K window with room to spare.
	messages := []Message{{Role: RoleUser, Content: "Hello"}}
	if !checker.Fits(messages, "gpt-4o", 0) {
		t.Error("expected short message to fit in gpt-4o window")
	}
}

func TestContextChecker_Fits_ExceedsWindow(t *testing.T) {
	checker := ContextChecker{}
	// Craft a message that exceeds the window: gpt-3.5-turbo has 16_385 tokens.
	// A string of 16_385*4+4 chars = 65_544 chars should estimate to 16_386 tokens.
	bigContent := make([]byte, 65_544)
	for i := range bigContent {
		bigContent[i] = 'a'
	}
	messages := []Message{{Role: RoleUser, Content: string(bigContent)}}
	if checker.Fits(messages, "gpt-3.5-turbo", 0) {
		t.Error("expected oversized message to not fit")
	}
}

func TestContextChecker_Fits_WithReserved(t *testing.T) {
	checker := ContextChecker{}
	messages := []Message{{Role: RoleUser, Content: "Hello"}}
	// Reserve almost the entire window — should still fit.
	if !checker.Fits(messages, "gpt-4o", 127_990) {
		t.Error("expected message to fit with large reserved")
	}
	// Reserve more than the entire window.
	if checker.Fits(messages, "gpt-4o", 128_001) {
		t.Error("expected message to not fit with reserved exceeding window")
	}
}

func TestContextChecker_Fits_UnknownModel(t *testing.T) {
	checker := ContextChecker{}
	messages := []Message{{Role: RoleUser, Content: "Hello"}}
	if checker.Fits(messages, "unknown-model", 0) {
		t.Error("expected unknown model to return false")
	}
}

func TestContextChecker_Remaining_Normal(t *testing.T) {
	checker := ContextChecker{}
	messages := []Message{{Role: RoleUser, Content: "Hello"}}
	rem, err := checker.Remaining(messages, "gpt-4o", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	estimated := EstimateTokens(messages)
	expected := 128_000 - estimated
	if rem != expected {
		t.Errorf("Remaining = %d, want %d", rem, expected)
	}
}

func TestContextChecker_Remaining_WithReserved(t *testing.T) {
	checker := ContextChecker{}
	messages := []Message{{Role: RoleUser, Content: "Hello"}}
	reserved := 1000
	rem, err := checker.Remaining(messages, "gpt-4o", reserved)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	estimated := EstimateTokens(messages)
	expected := 128_000 - estimated - reserved
	if rem != expected {
		t.Errorf("Remaining = %d, want %d", rem, expected)
	}
}

func TestContextChecker_Remaining_UnknownModel(t *testing.T) {
	checker := ContextChecker{}
	_, err := checker.Remaining([]Message{}, "unknown-model", 0)
	if err == nil {
		t.Error("expected error for unknown model")
	}
	if err != ErrUnknownModel {
		t.Errorf("error = %v, want ErrUnknownModel", err)
	}
}
