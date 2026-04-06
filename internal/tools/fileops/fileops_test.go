package fileops

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kw12121212/spec-coding-sdk/internal/core"
	"github.com/kw12121212/spec-coding-sdk/internal/permission"
)

// Compile-time checks that tools satisfy core.Tool.
var (
	_ core.Tool = (*ReadTool)(nil)
	_ core.Tool = (*WriteTool)(nil)
	_ core.Tool = (*EditTool)(nil)
)

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

func newPolicyProviderForDecision(operation string, decision permission.Decision) core.PermissionProvider {
	return permission.NewPolicyProvider(permission.NewStaticPolicy(
		permission.DecisionAllow,
		permission.Rule{OperationPattern: operation, Decision: decision},
	))
}

// --- ReadTool tests ---

func TestReadTool_NormalRead(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	content := "line1\nline2\nline3\n"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	tool := NewReadTool(nil)
	result, err := tool.Execute(context.Background(), mustMarshal(t, ReadInput{FilePath: tmpFile}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Output)
	}
	if !strings.Contains(result.Output, "1\tline1") {
		t.Fatalf("expected line-numbered output, got %q", result.Output)
	}
	if !strings.Contains(result.Output, "2\tline2") {
		t.Fatalf("expected line 2, got %q", result.Output)
	}
	if !strings.Contains(result.Output, "3\tline3") {
		t.Fatalf("expected line 3, got %q", result.Output)
	}
}

func TestReadTool_LineNumbers(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpFile, []byte("aaa\nbbb\n"), 0644); err != nil {
		t.Fatal(err)
	}

	tool := NewReadTool(nil)
	result, err := tool.Execute(context.Background(), mustMarshal(t, ReadInput{FilePath: tmpFile}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(result.Output), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	if !strings.HasPrefix(lines[0], "1\t") {
		t.Fatalf("expected line number prefix '1\\t', got %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], "2\t") {
		t.Fatalf("expected line number prefix '2\\t', got %q", lines[1])
	}
}

func TestReadTool_Offset(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpFile, []byte("a\nb\nc\nd\n"), 0644); err != nil {
		t.Fatal(err)
	}

	tool := NewReadTool(nil)
	result, err := tool.Execute(context.Background(), mustMarshal(t, ReadInput{FilePath: tmpFile, Offset: 3}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Output)
	}
	if !strings.Contains(result.Output, "3\tc") {
		t.Fatalf("expected line 3 (offset=3), got %q", result.Output)
	}
	if !strings.Contains(result.Output, "4\td") {
		t.Fatalf("expected line 4, got %q", result.Output)
	}
	if strings.Contains(result.Output, "1\ta") {
		t.Fatalf("should not contain line 1, got %q", result.Output)
	}
}

func TestReadTool_Limit(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpFile, []byte("a\nb\nc\nd\n"), 0644); err != nil {
		t.Fatal(err)
	}

	tool := NewReadTool(nil)
	result, err := tool.Execute(context.Background(), mustMarshal(t, ReadInput{FilePath: tmpFile, Limit: 2}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(result.Output), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines with limit=2, got %d: %q", len(lines), result.Output)
	}
}

func TestReadTool_FileNotExist(t *testing.T) {
	tool := NewReadTool(nil)
	result, err := tool.Execute(context.Background(), mustMarshal(t, ReadInput{FilePath: "/nonexistent/file.txt"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError for missing file")
	}
}

func TestReadTool_PathIsDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	tool := NewReadTool(nil)
	result, err := tool.Execute(context.Background(), mustMarshal(t, ReadInput{FilePath: tmpDir}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError when reading a directory")
	}
}

func TestReadTool_EmptyFilePath(t *testing.T) {
	tool := NewReadTool(nil)
	_, err := tool.Execute(context.Background(), mustMarshal(t, ReadInput{FilePath: ""}))
	if err == nil {
		t.Fatal("expected error for empty file_path")
	}
}

func TestReadTool_RelativePath(t *testing.T) {
	tool := NewReadTool(nil)
	_, err := tool.Execute(context.Background(), mustMarshal(t, ReadInput{FilePath: "relative/path.txt"}))
	if err == nil {
		t.Fatal("expected error for relative path")
	}
}

func TestReadTool_InvalidJSON(t *testing.T) {
	tool := NewReadTool(nil)
	_, err := tool.Execute(context.Background(), json.RawMessage(`{invalid`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

// --- WriteTool tests ---

func TestWriteTool_NormalWrite(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "output.txt")
	tool := NewWriteTool(nil)
	result, err := tool.Execute(context.Background(), mustMarshal(t, WriteInput{
		FilePath: tmpFile,
		Content:  "hello world",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Output)
	}

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello world" {
		t.Fatalf("expected 'hello world', got %q", string(data))
	}
}

func TestWriteTool_AutoCreateParentDir(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "a", "b", "c", "deep.txt")
	tool := NewWriteTool(nil)
	result, err := tool.Execute(context.Background(), mustMarshal(t, WriteInput{
		FilePath: filePath,
		Content:  "nested",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Output)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "nested" {
		t.Fatalf("expected 'nested', got %q", string(data))
	}
}

func TestWriteTool_OverwriteExisting(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(tmpFile, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}

	tool := NewWriteTool(nil)
	result, err := tool.Execute(context.Background(), mustMarshal(t, WriteInput{
		FilePath: tmpFile,
		Content:  "new",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Output)
	}

	data, _ := os.ReadFile(tmpFile)
	if string(data) != "new" {
		t.Fatalf("expected 'new', got %q", string(data))
	}
}

func TestWriteTool_EmptyContent(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "empty.txt")
	tool := NewWriteTool(nil)
	result, err := tool.Execute(context.Background(), mustMarshal(t, WriteInput{
		FilePath: tmpFile,
		Content:  "",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Output)
	}

	data, _ := os.ReadFile(tmpFile)
	if string(data) != "" {
		t.Fatalf("expected empty file, got %q", string(data))
	}
}

func TestWriteTool_EmptyFilePath(t *testing.T) {
	tool := NewWriteTool(nil)
	_, err := tool.Execute(context.Background(), mustMarshal(t, WriteInput{FilePath: ""}))
	if err == nil {
		t.Fatal("expected error for empty file_path")
	}
}

func TestWriteTool_RelativePath(t *testing.T) {
	tool := NewWriteTool(nil)
	_, err := tool.Execute(context.Background(), mustMarshal(t, WriteInput{FilePath: "relative.txt"}))
	if err == nil {
		t.Fatal("expected error for relative path")
	}
}

func TestWriteTool_InvalidJSON(t *testing.T) {
	tool := NewWriteTool(nil)
	_, err := tool.Execute(context.Background(), json.RawMessage(`{invalid`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

// --- EditTool tests ---

func TestEditTool_SingleReplace(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "edit.txt")
	if err := os.WriteFile(tmpFile, []byte("hello world"), 0644); err != nil {
		t.Fatal(err)
	}

	tool := NewEditTool(nil)
	result, err := tool.Execute(context.Background(), mustMarshal(t, EditInput{
		FilePath:  tmpFile,
		OldString: "world",
		NewString: "Go",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Output)
	}

	data, _ := os.ReadFile(tmpFile)
	if string(data) != "hello Go" {
		t.Fatalf("expected 'hello Go', got %q", string(data))
	}
}

func TestEditTool_MultipleMatchesError(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "edit.txt")
	if err := os.WriteFile(tmpFile, []byte("aaa bbb aaa"), 0644); err != nil {
		t.Fatal(err)
	}

	tool := NewEditTool(nil)
	result, err := tool.Execute(context.Background(), mustMarshal(t, EditInput{
		FilePath:  tmpFile,
		OldString: "aaa",
		NewString: "zzz",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for multiple matches")
	}
	if !strings.Contains(result.Output, "2 locations") {
		t.Fatalf("expected multiple match error, got %q", result.Output)
	}

	// Verify file unchanged
	data, _ := os.ReadFile(tmpFile)
	if string(data) != "aaa bbb aaa" {
		t.Fatalf("file should be unchanged, got %q", string(data))
	}
}

func TestEditTool_NotFoundError(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "edit.txt")
	if err := os.WriteFile(tmpFile, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	tool := NewEditTool(nil)
	result, err := tool.Execute(context.Background(), mustMarshal(t, EditInput{
		FilePath:  tmpFile,
		OldString: "missing",
		NewString: "replacement",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected error for not found")
	}
}

func TestEditTool_ReplaceAll(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "edit.txt")
	if err := os.WriteFile(tmpFile, []byte("aaa bbb aaa ccc aaa"), 0644); err != nil {
		t.Fatal(err)
	}

	tool := NewEditTool(nil)
	result, err := tool.Execute(context.Background(), mustMarshal(t, EditInput{
		FilePath:   tmpFile,
		OldString:  "aaa",
		NewString:  "xxx",
		ReplaceAll: true,
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Output)
	}
	if !strings.Contains(result.Output, "3 occurrences") {
		t.Fatalf("expected '3 occurrences', got %q", result.Output)
	}

	data, _ := os.ReadFile(tmpFile)
	if string(data) != "xxx bbb xxx ccc xxx" {
		t.Fatalf("expected all replaced, got %q", string(data))
	}
}

func TestEditTool_EmptyFilePath(t *testing.T) {
	tool := NewEditTool(nil)
	_, err := tool.Execute(context.Background(), mustMarshal(t, EditInput{
		FilePath:  "",
		OldString: "x",
		NewString: "y",
	}))
	if err == nil {
		t.Fatal("expected error for empty file_path")
	}
}

func TestEditTool_EmptyOldString(t *testing.T) {
	tool := NewEditTool(nil)
	_, err := tool.Execute(context.Background(), mustMarshal(t, EditInput{
		FilePath:  "/tmp/test.txt",
		OldString: "",
		NewString: "y",
	}))
	if err == nil {
		t.Fatal("expected error for empty old_string")
	}
}

func TestEditTool_RelativePath(t *testing.T) {
	tool := NewEditTool(nil)
	_, err := tool.Execute(context.Background(), mustMarshal(t, EditInput{
		FilePath:  "relative.txt",
		OldString: "x",
		NewString: "y",
	}))
	if err == nil {
		t.Fatal("expected error for relative path")
	}
}

func TestEditTool_InvalidJSON(t *testing.T) {
	tool := NewEditTool(nil)
	_, err := tool.Execute(context.Background(), json.RawMessage(`{invalid`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestEditTool_FileNotExist(t *testing.T) {
	tool := NewEditTool(nil)
	result, err := tool.Execute(context.Background(), mustMarshal(t, EditInput{
		FilePath:  "/nonexistent/file.txt",
		OldString: "x",
		NewString: "y",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError for missing file")
	}
}

// --- Permission tests ---

func TestReadTool_PermissionDenied(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpFile, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	tool := NewReadTool(&mockPermissionProvider{err: fmt.Errorf("denied")})
	result, err := tool.Execute(context.Background(), mustMarshal(t, ReadInput{FilePath: tmpFile}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError for permission denied")
	}
}

func TestReadTool_PermissionNil(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpFile, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	tool := NewReadTool(nil)
	result, err := tool.Execute(context.Background(), mustMarshal(t, ReadInput{FilePath: tmpFile}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Output)
	}
}

func TestWriteTool_PermissionDenied(t *testing.T) {
	tool := NewWriteTool(&mockPermissionProvider{err: fmt.Errorf("denied")})
	result, err := tool.Execute(context.Background(), mustMarshal(t, WriteInput{
		FilePath: "/tmp/test.txt",
		Content:  "x",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError for permission denied")
	}
}

func TestWriteTool_PermissionDeniedDoesNotCreateFile(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "blocked.txt")
	tool := NewWriteTool(newPolicyProviderForDecision("file:write", permission.DecisionDeny))

	result, err := tool.Execute(context.Background(), mustMarshal(t, WriteInput{
		FilePath: target,
		Content:  "blocked",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError for permission denied")
	}
	if result.Output != "permission denied: file:write on "+target {
		t.Fatalf("unexpected permission error: %q", result.Output)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("expected file to be absent, stat err=%v", err)
	}
}

func TestWriteTool_NeedConfirmationDoesNotCreateFile(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "blocked.txt")
	tool := NewWriteTool(newPolicyProviderForDecision("file:write", permission.DecisionNeedConfirmation))

	result, err := tool.Execute(context.Background(), mustMarshal(t, WriteInput{
		FilePath: target,
		Content:  "blocked",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError for need confirmation")
	}
	if result.Output != "confirmation required: file:write on "+target {
		t.Fatalf("unexpected confirmation error: %q", result.Output)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("expected file to be absent, stat err=%v", err)
	}
}

func TestWriteTool_PermissionNil(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	tool := NewWriteTool(nil)
	result, err := tool.Execute(context.Background(), mustMarshal(t, WriteInput{
		FilePath: tmpFile,
		Content:  "x",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Output)
	}
}

func TestEditTool_PermissionDenied(t *testing.T) {
	tool := NewEditTool(&mockPermissionProvider{err: fmt.Errorf("denied")})
	result, err := tool.Execute(context.Background(), mustMarshal(t, EditInput{
		FilePath:  "/tmp/test.txt",
		OldString: "x",
		NewString: "y",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError for permission denied")
	}
}

func TestEditTool_PermissionDeniedDoesNotModifyFile(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "edit.txt")
	if err := os.WriteFile(tmpFile, []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}

	tool := NewEditTool(newPolicyProviderForDecision("file:edit", permission.DecisionDeny))
	result, err := tool.Execute(context.Background(), mustMarshal(t, EditInput{
		FilePath:  tmpFile,
		OldString: "original",
		NewString: "changed",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError for permission denied")
	}
	if result.Output != "permission denied: file:edit on "+tmpFile {
		t.Fatalf("unexpected permission error: %q", result.Output)
	}

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "original" {
		t.Fatalf("expected file to remain unchanged, got %q", string(data))
	}
}

func TestEditTool_NeedConfirmationDoesNotModifyFile(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "edit.txt")
	if err := os.WriteFile(tmpFile, []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}

	tool := NewEditTool(newPolicyProviderForDecision("file:edit", permission.DecisionNeedConfirmation))
	result, err := tool.Execute(context.Background(), mustMarshal(t, EditInput{
		FilePath:  tmpFile,
		OldString: "original",
		NewString: "changed",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError for need confirmation")
	}
	if result.Output != "confirmation required: file:edit on "+tmpFile {
		t.Fatalf("unexpected confirmation error: %q", result.Output)
	}

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "original" {
		t.Fatalf("expected file to remain unchanged, got %q", string(data))
	}
}

func TestEditTool_PermissionNil(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpFile, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	tool := NewEditTool(nil)
	result, err := tool.Execute(context.Background(), mustMarshal(t, EditInput{
		FilePath:  tmpFile,
		OldString: "hello",
		NewString: "world",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error: %s", result.Output)
	}
}
