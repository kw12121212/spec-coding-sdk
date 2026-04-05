package glob

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// mockPerms is a test permission provider that denies specific operations.
type mockPerms struct {
	denied map[string]bool
}

func (m *mockPerms) Check(_ context.Context, operation string, _ string) error {
	if m.denied[operation] {
		return fmt.Errorf("permission denied: %s", operation)
	}
	return nil
}

// createTestFiles creates a temp dir with predictable file structure.
func createTestFiles(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Create files with different modification times
	files := []struct {
		path    string
		content string
	}{
		{"a.txt", "aaa"},
		{"b.go", "bbb"},
		{"sub/c.txt", "ccc"},
		{"sub/d.go", "ddd"},
		{"sub/deep/e.txt", "eee"},
	}
	for _, f := range files {
		full := filepath.Join(dir, f.path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(f.content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestBasicGlobPattern(t *testing.T) {
	dir := createTestFiles(t)
	tool := New(nil)

	raw, _ := json.Marshal(Input{Pattern: "*.txt", Path: dir})
	result, err := tool.Execute(context.Background(), raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error result: %s", result.Output)
	}
	if !containsLine(result.Output, filepath.Join(dir, "a.txt")) {
		t.Errorf("expected a.txt in output, got: %s", result.Output)
	}
	if containsLine(result.Output, filepath.Join(dir, "b.go")) {
		t.Errorf("did not expect b.go in output, got: %s", result.Output)
	}
}

func TestRecursiveDoublestar(t *testing.T) {
	dir := createTestFiles(t)
	tool := New(nil)

	raw, _ := json.Marshal(Input{Pattern: "**/*.txt", Path: dir})
	result, err := tool.Execute(context.Background(), raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error result: %s", result.Output)
	}

	expected := []string{
		filepath.Join(dir, "a.txt"),
		filepath.Join(dir, "sub", "c.txt"),
		filepath.Join(dir, "sub", "deep", "e.txt"),
	}
	for _, exp := range expected {
		if !containsLine(result.Output, exp) {
			t.Errorf("expected %s in output, got: %s", exp, result.Output)
		}
	}
	// Should not include .go files
	if containsLine(result.Output, ".go") {
		t.Errorf("should not include .go files, got: %s", result.Output)
	}
}

func TestEmptyPatternReturnsError(t *testing.T) {
	tool := New(nil)
	raw, _ := json.Marshal(Input{Pattern: ""})
	_, err := tool.Execute(context.Background(), raw)
	if err == nil {
		t.Fatal("expected error for empty pattern")
	}
}

func TestNonExistentPathReturnsError(t *testing.T) {
	tool := New(nil)
	raw, _ := json.Marshal(Input{Pattern: "*.txt", Path: "/nonexistent/path"})
	result, err := tool.Execute(context.Background(), raw)
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result for non-existent path")
	}
}

func TestPathIsFileReturnsError(t *testing.T) {
	dir := createTestFiles(t)
	filePath := filepath.Join(dir, "a.txt")
	tool := New(nil)
	raw, _ := json.Marshal(Input{Pattern: "*.txt", Path: filePath})
	result, err := tool.Execute(context.Background(), raw)
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result when path is a file")
	}
}

func TestPermissionDenied(t *testing.T) {
	dir := createTestFiles(t)
	perms := &mockPerms{denied: map[string]bool{"glob:execute": true}}
	tool := New(perms)

	raw, _ := json.Marshal(Input{Pattern: "*.txt", Path: dir})
	result, err := tool.Execute(context.Background(), raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error result when permission denied")
	}
}

func TestPermissionNilSkipsCheck(t *testing.T) {
	dir := createTestFiles(t)
	tool := New(nil)

	raw, _ := json.Marshal(Input{Pattern: "*.txt", Path: dir})
	result, err := tool.Execute(context.Background(), raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Output)
	}
	if result.Output == "" {
		t.Fatal("expected non-empty output")
	}
}

func TestOutputTruncation(t *testing.T) {
	dir := t.TempDir()
	// Create many files to exceed 1MB of output paths
	for i := 0; i < 50000; i++ {
		name := fmt.Sprintf("file_with_a_very_long_name_to_fill_output_buffer_%d.txt", i)
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	tool := New(nil)
	raw, _ := json.Marshal(Input{Pattern: "*.txt", Path: dir})
	result, err := tool.Execute(context.Background(), raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected truncation error")
	}
	if len(result.Output) > maxOutputBytes+len(truncationNotice) {
		t.Errorf("output too large: %d bytes", len(result.Output))
	}
}

func TestSortByModTime(t *testing.T) {
	dir := t.TempDir()

	f1 := filepath.Join(dir, "first.txt")
	f2 := filepath.Join(dir, "second.txt")
	f3 := filepath.Join(dir, "third.txt")

	os.WriteFile(f1, []byte("1"), 0o644)
	time.Sleep(10 * time.Millisecond)
	os.WriteFile(f2, []byte("2"), 0o644)
	time.Sleep(10 * time.Millisecond)
	os.WriteFile(f3, []byte("3"), 0o644)

	tool := New(nil)
	raw, _ := json.Marshal(Input{Pattern: "*.txt", Path: dir})
	result, err := tool.Execute(context.Background(), raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Output)
	}

	// Third should appear before second, second before first (most recent first)
	lines := splitLines(result.Output)
	if len(lines) < 3 {
		t.Fatalf("expected at least 3 lines, got: %s", result.Output)
	}
	if lines[0] != f3 {
		t.Errorf("expected first line to be %s, got %s", f3, lines[0])
	}
	if lines[1] != f2 {
		t.Errorf("expected second line to be %s, got %s", f2, lines[1])
	}
	if lines[2] != f1 {
		t.Errorf("expected third line to be %s, got %s", f1, lines[2])
	}
}

func TestDefaultPathIsCWD(t *testing.T) {
	dir := createTestFiles(t)
	tool := New(nil)

	// Change working directory to test dir
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	raw, _ := json.Marshal(Input{Pattern: "*.txt"})
	result, err := tool.Execute(context.Background(), raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Output)
	}
	if !containsLine(result.Output, filepath.Join(dir, "a.txt")) {
		t.Errorf("expected a.txt in output, got: %s", result.Output)
	}
}

func TestInvalidJSON(t *testing.T) {
	tool := New(nil)
	_, err := tool.Execute(context.Background(), json.RawMessage(`{invalid`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func containsLine(output, line string) bool {
	for _, l := range splitLines(output) {
		if l == line {
			return true
		}
	}
	return false
}

func splitLines(s string) []string {
	return splitNonEmpty(strings.Split(s, "\n"))
}

func splitNonEmpty(parts []string) []string {
	var result []string
	for _, p := range parts {
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
