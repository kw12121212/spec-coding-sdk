package grep

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kw12121212/spec-coding-sdk/internal/core"
)

// Compile-time check that Tool satisfies core.Tool.
var _ core.Tool = (*Tool)(nil)

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

// skipIfNoRg skips the test if rg is not available in PATH.
func skipIfNoRg(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("rg"); err != nil {
		t.Skip("rg not found in PATH, skipping test")
	}
}

// createTestFiles creates a temp dir with sample files for search.
func createTestFiles(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "hello.go"), []byte("package main\n\nfunc main() {\n\tfmt.Println(\"hello world\")\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("Hello World\nThis is a test file.\nGoodbye world\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "data.py"), []byte("def hello():\n    print('hello from python')\n    return True\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestExecute_HappyPath(t *testing.T) {
	skipIfNoRg(t)
	dir := createTestFiles(t)

	tool := New(nil)
	input := mustMarshal(t, Input{Pattern: "hello", Path: dir})

	result, err := tool.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error result: %s", result.Output)
	}
	if !strings.Contains(strings.ToLower(result.Output), "hello") {
		t.Fatalf("expected 'hello' in output, got %q", result.Output)
	}
}

func TestExecute_NoMatch(t *testing.T) {
	skipIfNoRg(t)
	dir := createTestFiles(t)

	tool := New(nil)
	input := mustMarshal(t, Input{Pattern: "zzz_nonexistent_pattern", Path: dir})

	result, err := tool.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("no-match should not be an error, got: %s", result.Output)
	}
	if !strings.Contains(result.Output, "no matches found") {
		t.Fatalf("expected no-match hint, got %q", result.Output)
	}
}

func TestExecute_EmptyPattern(t *testing.T) {
	tool := New(nil)
	input := mustMarshal(t, Input{Pattern: ""})

	_, err := tool.Execute(context.Background(), input)
	if err == nil {
		t.Fatal("expected error for empty pattern")
	}
}

func TestExecute_InvalidJSON(t *testing.T) {
	tool := New(nil)

	_, err := tool.Execute(context.Background(), json.RawMessage(`{invalid`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestExecute_OutputMode_FilesWithMatches(t *testing.T) {
	skipIfNoRg(t)
	dir := createTestFiles(t)

	tool := New(nil)
	input := mustMarshal(t, Input{Pattern: "hello", Path: dir, OutputMode: "files_with_matches"})

	result, err := tool.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Output)
	}
	lines := strings.Split(strings.TrimSpace(result.Output), "\n")
	if len(lines) < 1 {
		t.Fatalf("expected at least one file, got %q", result.Output)
	}
	// Each line should be a file path, not contain line numbers
	for _, line := range lines {
		if strings.Contains(line, ":") {
			t.Fatalf("files_with_matches should not contain colons, got %q", line)
		}
	}
}

func TestExecute_OutputMode_Count(t *testing.T) {
	skipIfNoRg(t)
	dir := createTestFiles(t)

	tool := New(nil)
	input := mustMarshal(t, Input{Pattern: "hello", Path: dir, OutputMode: "count"})

	result, err := tool.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Output)
	}
	// Output should contain counts (digits after colon)
	if !strings.Contains(result.Output, ":") {
		t.Fatalf("count mode should contain file:count pairs, got %q", result.Output)
	}
}

func TestExecute_IgnoreCase(t *testing.T) {
	skipIfNoRg(t)
	dir := createTestFiles(t)

	tool := New(nil)
	input := mustMarshal(t, Input{Pattern: "HELLO", Path: dir, IgnoreCase: true})

	result, err := tool.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Output)
	}
	if !strings.Contains(strings.ToLower(result.Output), "hello") {
		t.Fatalf("expected match with ignore_case, got %q", result.Output)
	}
}

func TestExecute_Glob(t *testing.T) {
	skipIfNoRg(t)
	dir := createTestFiles(t)

	tool := New(nil)
	input := mustMarshal(t, Input{Pattern: "hello", Path: dir, Glob: "*.go"})

	result, err := tool.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Output)
	}
	if !strings.Contains(result.Output, "hello.go") {
		t.Fatalf("expected hello.go in results, got %q", result.Output)
	}
	if strings.Contains(result.Output, "readme.txt") {
		t.Fatalf("glob *.go should exclude readme.txt, got %q", result.Output)
	}
}

func TestExecute_Type(t *testing.T) {
	skipIfNoRg(t)
	dir := createTestFiles(t)

	tool := New(nil)
	input := mustMarshal(t, Input{Pattern: "hello", Path: dir, Type: "go"})

	result, err := tool.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Output)
	}
	if !strings.Contains(result.Output, "hello.go") {
		t.Fatalf("expected hello.go in results, got %q", result.Output)
	}
}

func TestExecute_Context(t *testing.T) {
	skipIfNoRg(t)
	dir := createTestFiles(t)

	tool := New(nil)
	input := mustMarshal(t, Input{Pattern: "fmt.Println", Path: dir, Context: 1})

	result, err := tool.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Output)
	}
	// With context=1, we should see surrounding lines
	if !strings.Contains(result.Output, "func main()") {
		t.Fatalf("expected context line 'func main()' in output, got %q", result.Output)
	}
}

func TestExecute_HeadLimit(t *testing.T) {
	skipIfNoRg(t)
	dir := createTestFiles(t)

	tool := New(nil)
	input := mustMarshal(t, Input{Pattern: "hello", Path: dir, HeadLimit: 1})

	result, err := tool.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Output)
	}
	lines := strings.Split(strings.TrimSpace(result.Output), "\n")
	if len(lines) > 1 {
		t.Fatalf("head_limit=1 should return at most 1 line, got %d lines", len(lines))
	}
}

func TestExecute_OutputTruncation(t *testing.T) {
	skipIfNoRg(t)
	// Create a file with >1MB content
	dir := t.TempDir()
	largeContent := strings.Repeat("x match line\n", 100000) // ~1.3MB
	if err := os.WriteFile(filepath.Join(dir, "big.txt"), []byte(largeContent), 0o644); err != nil {
		t.Fatal(err)
	}

	tool := New(nil)
	input := mustMarshal(t, Input{Pattern: "match", Path: dir})

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
	if len(result.Output) > maxOutputBytes+200 {
		t.Fatalf("output too large: %d bytes", len(result.Output))
	}
}

func TestExecute_RgNotFound(t *testing.T) {
	// Temporarily override PATH to ensure rg is not found
	origPath := os.Getenv("PATH")
	t.Cleanup(func() { _ = os.Setenv("PATH", origPath) })

	tmpDir := t.TempDir()
	_ = os.Setenv("PATH", tmpDir)

	tool := New(nil)
	input := mustMarshal(t, Input{Pattern: "test"})

	result, err := tool.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError=true when rg not found")
	}
	if !strings.Contains(result.Output, "rg not found") {
		t.Fatalf("expected 'rg not found' message, got %q", result.Output)
	}
}

func TestExecute_PermissionDenied(t *testing.T) {
	skipIfNoRg(t)
	dir := createTestFiles(t)

	tool := New(&mockPermissionProvider{
		err: fmt.Errorf("permission denied: restricted pattern"),
	})
	input := mustMarshal(t, Input{Pattern: "secret", Path: dir})

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
	skipIfNoRg(t)
	dir := createTestFiles(t)

	tool := New(nil)
	input := mustMarshal(t, Input{Pattern: "hello", Path: dir})

	result, err := tool.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Output)
	}
}
