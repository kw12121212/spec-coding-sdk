package bash

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kw12121212/spec-coding-sdk/internal/core"
)

// Compile-time check that Tool satisfies core.Tool.
var _ core.Tool = (*Tool)(nil)

// mockPermissionProvider is a test-only PermissionProvider.
type mockPermissionProvider struct {
	err error
}

func (m *mockPermissionProvider) Check(_ context.Context, _, _ string) error {
	return m.err
}

func mustMarshal(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal input: %v", err)
	}
	return b
}

func TestExecute_HappyPath(t *testing.T) {
	tool := New(nil)
	input := mustMarshal(t, Input{Command: "echo hello"})

	result, err := tool.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error result: %s", result.Output)
	}
	got := strings.TrimSpace(result.Output)
	if got != "hello" {
		t.Fatalf("expected output 'hello', got %q", got)
	}
}

func TestExecute_ShellExpression(t *testing.T) {
	tool := New(nil)
	input := mustMarshal(t, Input{Command: "echo $((1+2))"})

	result, err := tool.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error result: %s", result.Output)
	}
	got := strings.TrimSpace(result.Output)
	if got != "3" {
		t.Fatalf("expected '3', got %q", got)
	}
}

func TestExecute_StderrMerged(t *testing.T) {
	tool := New(nil)
	input := mustMarshal(t, Input{Command: "echo out && echo err >&2"})

	result, err := tool.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result.Output, "out") {
		t.Fatalf("expected 'out' in output, got %q", result.Output)
	}
	if !strings.Contains(result.Output, "err") {
		t.Fatalf("expected 'err' in output, got %q", result.Output)
	}
}

func TestExecute_Timeout(t *testing.T) {
	tool := New(nil)
	input := mustMarshal(t, Input{Command: "sleep 30", Timeout: 1})

	result, err := tool.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError=true for timeout")
	}
	if !strings.Contains(result.Output, "timed out") {
		t.Fatalf("expected timeout message, got %q", result.Output)
	}
}

func TestExecute_OutputTruncation(t *testing.T) {
	tool := New(nil)
	// Generate >1MB output
	input := mustMarshal(t, Input{Command: "python3 -c \"print('x'*1100000)\""})

	result, err := tool.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError=true for truncated output")
	}
	if !strings.Contains(result.Output, "truncated") {
		t.Fatalf("expected truncation notice, got output of %d bytes", len(result.Output))
	}
	if len(result.Output) > 1<<20+200 { // 1MB + notice text
		t.Fatalf("output too large: %d bytes", len(result.Output))
	}
}

func TestExecute_PermissionDenied(t *testing.T) {
	tool := New(&mockPermissionProvider{
		err: fmt.Errorf("permission denied: dangerous command"),
	})
	input := mustMarshal(t, Input{Command: "rm -rf /"})

	result, err := tool.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError=true for permission denied")
	}
	if !strings.Contains(result.Output, "permission denied") {
		t.Fatalf("expected permission denied message, got %q", result.Output)
	}
}

func TestExecute_PermissionNil(t *testing.T) {
	tool := New(nil)
	input := mustMarshal(t, Input{Command: "echo works"})

	result, err := tool.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Output)
	}
}

func TestExecute_WorkingDir(t *testing.T) {
	tmpDir := t.TempDir()
	tool := New(nil)
	input := mustMarshal(t, Input{Command: "pwd", WorkingDir: tmpDir})

	result, err := tool.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Output)
	}

	got := strings.TrimSpace(result.Output)
	absTmp, _ := filepath.Abs(tmpDir)
	if got != absTmp {
		t.Fatalf("expected working dir %q, got %q", absTmp, got)
	}
}

func TestExecute_EmptyCommand(t *testing.T) {
	tool := New(nil)
	input := mustMarshal(t, Input{Command: ""})

	_, err := tool.Execute(context.Background(), input)
	if err == nil {
		t.Fatal("expected error for empty command")
	}
}

func TestExecute_InvalidJSON(t *testing.T) {
	tool := New(nil)

	_, err := tool.Execute(context.Background(), json.RawMessage(`{invalid`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestExecute_NonZeroExitCode(t *testing.T) {
	tool := New(nil)
	input := mustMarshal(t, Input{Command: "exit 42"})

	result, err := tool.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError=true for non-zero exit code")
	}
}

func TestExecute_CommandNotFound(t *testing.T) {
	tool := New(nil)
	input := mustMarshal(t, Input{Command: "nonexistent_command_xyz"})

	result, err := tool.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError=true for command not found")
	}
}

func TestExecute_EnvironmentInherited(t *testing.T) {
	t.Setenv("SPEC_TEST_VAR", "testvalue")

	tool := New(nil)
	input := mustMarshal(t, Input{Command: "echo $SPEC_TEST_VAR"})

	result, err := tool.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Output)
	}
	got := strings.TrimSpace(result.Output)
	if got != "testvalue" {
		t.Fatalf("expected 'testvalue', got %q", got)
	}
}
