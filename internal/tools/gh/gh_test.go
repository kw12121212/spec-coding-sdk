package gh

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kw12121212/spec-coding-sdk/internal/core"
	"github.com/kw12121212/spec-coding-sdk/internal/permission"
)

// Compile-time check that Tool satisfies core.Tool.
var _ core.Tool = (*Tool)(nil)

type fakeResolver struct {
	result resolvedExecutable
	err    error
	calls  int
}

func (f *fakeResolver) Resolve(_ context.Context, _ string) (resolvedExecutable, error) {
	f.calls++
	if f.err != nil {
		return resolvedExecutable{}, f.err
	}

	return f.result, nil
}

func mustMarshal(t *testing.T, v any) json.RawMessage {
	t.Helper()

	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal input: %v", err)
	}

	return data
}

func newPolicyProviderForDecision(operation string, decision permission.Decision) core.PermissionProvider {
	return permission.NewPolicyProvider(permission.NewStaticPolicy(
		permission.DecisionAllow,
		permission.Rule{OperationPattern: operation, Decision: decision},
	))
}

func writeExecutable(t *testing.T, dir, name, content string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("write executable: %v", err)
	}

	return path
}

func newTestTool(perms core.PermissionProvider, resolver executableResolver) *Tool {
	return &Tool{
		perms:    perms,
		resolver: resolver,
	}
}

func TestExecute_SystemInstalledHappyPath(t *testing.T) {
	tmpDir := t.TempDir()
	executable := writeExecutable(t, tmpDir, "gh-system", "#!/bin/sh\nprintf 'args:%s|%s\\n' \"$1\" \"$2\"\npwd\n")
	resolver := &fakeResolver{result: resolvedExecutable{Path: executable}}
	tool := newTestTool(nil, resolver)

	result, err := tool.Execute(context.Background(), mustMarshal(t, Input{
		Args:       []string{"repo", "view"},
		WorkingDir: tmpDir,
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error result: %s", result.Output)
	}
	if !strings.Contains(result.Output, "args:repo|view") {
		t.Fatalf("expected args in output, got %q", result.Output)
	}
	if !strings.Contains(result.Output, tmpDir) {
		t.Fatalf("expected working dir in output, got %q", result.Output)
	}
	if resolver.calls != 1 {
		t.Fatalf("expected resolver to be called once, got %d", resolver.calls)
	}
}

func TestExecute_ManagedInstallationFallback(t *testing.T) {
	tmpDir := t.TempDir()
	executable := writeExecutable(t, tmpDir, "gh-managed", "#!/bin/sh\nprintf 'managed:%s\\n' \"$1\"\n")
	resolver := &fakeResolver{result: resolvedExecutable{Path: executable}}
	tool := newTestTool(nil, resolver)

	result, err := tool.Execute(context.Background(), mustMarshal(t, Input{
		Args: []string{"auth", "status"},
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected error result: %s", result.Output)
	}
	if !strings.Contains(result.Output, "managed:auth") {
		t.Fatalf("expected managed executable output, got %q", result.Output)
	}
}

func TestExecute_PermissionDeniedDoesNotResolveOrRun(t *testing.T) {
	tmpDir := t.TempDir()
	executable := writeExecutable(t, tmpDir, "gh-blocked", "#!/bin/sh\necho should-not-run\n")
	resolver := &fakeResolver{result: resolvedExecutable{Path: executable}}
	tool := newTestTool(newPolicyProviderForDecision("gh:execute", permission.DecisionDeny), resolver)

	result, err := tool.Execute(context.Background(), mustMarshal(t, Input{
		Args: []string{"repo", "list"},
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected IsError=true for permission denied")
	}
	if result.Output != "permission denied: gh:execute on repo list" {
		t.Fatalf("unexpected permission error: %q", result.Output)
	}
	if resolver.calls != 0 {
		t.Fatalf("expected resolver not to be called, got %d", resolver.calls)
	}
}

func TestExecute_InvalidInput(t *testing.T) {
	t.Run("empty args", func(t *testing.T) {
		tool := New(nil)
		_, err := tool.Execute(context.Background(), mustMarshal(t, Input{}))
		if err == nil {
			t.Fatal("expected error for empty args")
		}
	})

	t.Run("empty arg element", func(t *testing.T) {
		tool := New(nil)
		_, err := tool.Execute(context.Background(), mustMarshal(t, Input{Args: []string{"repo", ""}}))
		if err == nil {
			t.Fatal("expected error for empty arg")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		tool := New(nil)
		_, err := tool.Execute(context.Background(), json.RawMessage(`{invalid`))
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})
}

func TestExecute_CommandFailureScenarios(t *testing.T) {
	t.Run("non-zero exit", func(t *testing.T) {
		tmpDir := t.TempDir()
		executable := writeExecutable(t, tmpDir, "gh-fail", "#!/bin/sh\necho failure >&2\nexit 7\n")
		tool := newTestTool(nil, &fakeResolver{result: resolvedExecutable{Path: executable}})

		result, err := tool.Execute(context.Background(), mustMarshal(t, Input{
			Args: []string{"repo", "delete"},
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Fatal("expected IsError=true for non-zero exit")
		}
		if !strings.Contains(result.Output, "failure") {
			t.Fatalf("expected command output in error result, got %q", result.Output)
		}
	})

	t.Run("timeout", func(t *testing.T) {
		tmpDir := t.TempDir()
		executable := writeExecutable(t, tmpDir, "gh-timeout", "#!/bin/sh\nexec sleep 30\n")
		tool := newTestTool(nil, &fakeResolver{result: resolvedExecutable{Path: executable}})

		result, err := tool.Execute(context.Background(), mustMarshal(t, Input{
			Args:    []string{"repo", "clone"},
			Timeout: 1,
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Fatal("expected IsError=true for timeout")
		}
		if !strings.Contains(result.Output, "timed out") {
			t.Fatalf("expected timeout message, got %q", result.Output)
		}
	})

	t.Run("output truncation", func(t *testing.T) {
		tmpDir := t.TempDir()
		executable := writeExecutable(t, tmpDir, "gh-truncate", "#!/bin/sh\npython3 -c \"print('x'*1100000)\"\n")
		tool := newTestTool(nil, &fakeResolver{result: resolvedExecutable{Path: executable}})

		result, err := tool.Execute(context.Background(), mustMarshal(t, Input{
			Args: []string{"repo", "view"},
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Fatal("expected IsError=true for truncated output")
		}
		if !strings.Contains(result.Output, "truncated") {
			t.Fatalf("expected truncation notice, got %q", result.Output)
		}
	})

	t.Run("resolve failure", func(t *testing.T) {
		tool := newTestTool(nil, &fakeResolver{err: errors.New("resolve failed")})

		result, err := tool.Execute(context.Background(), mustMarshal(t, Input{
			Args: []string{"repo", "view"},
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsError {
			t.Fatal("expected IsError=true for resolve failure")
		}
		if result.Output != "resolve failed" {
			t.Fatalf("unexpected resolve error: %q", result.Output)
		}
	})
}
