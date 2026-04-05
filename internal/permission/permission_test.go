package permission

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// --- Decision tests ---

func TestDecisionString(t *testing.T) {
	tests := []struct {
		d    Decision
		want string
	}{
		{DecisionAllow, "allow"},
		{DecisionDeny, "deny"},
		{DecisionNeedConfirmation, "need_confirmation"},
	}
	for _, tt := range tests {
		if got := tt.d.String(); got != tt.want {
			t.Errorf("Decision(%q).String() = %q, want %q", tt.d, got, tt.want)
		}
	}
}

func TestDecisionConstants(t *testing.T) {
	if DecisionAllow != "allow" {
		t.Errorf("DecisionAllow = %q, want %q", DecisionAllow, "allow")
	}
	if DecisionDeny != "deny" {
		t.Errorf("DecisionDeny = %q, want %q", DecisionDeny, "deny")
	}
	if DecisionNeedConfirmation != "need_confirmation" {
		t.Errorf("DecisionNeedConfirmation = %q, want %q", DecisionNeedConfirmation, "need_confirmation")
	}
}

// --- Rule.Match tests ---

func TestRuleMatch_ExactOperation(t *testing.T) {
	r := Rule{OperationPattern: "bash:execute", Decision: DecisionAllow}
	if !r.Match("bash:execute", "ls") {
		t.Error("expected exact operation match")
	}
	if r.Match("file:read", "ls") {
		t.Error("should not match different operation")
	}
}

func TestRuleMatch_WildcardOperation(t *testing.T) {
	r := Rule{OperationPattern: "file:*", Decision: DecisionAllow}
	if !r.Match("file:read", "/tmp/x") {
		t.Error("expected wildcard match on file:read")
	}
	if !r.Match("file:write", "/tmp/x") {
		t.Error("expected wildcard match on file:write")
	}
	if r.Match("bash:execute", "ls") {
		t.Error("should not match bash:execute with file:* pattern")
	}
}

func TestRuleMatch_EmptyResourcePattern(t *testing.T) {
	r := Rule{OperationPattern: "bash:*", ResourcePattern: "", Decision: DecisionDeny}
	if !r.Match("bash:execute", "anything") {
		t.Error("empty ResourcePattern should match all resources")
	}
}

func TestRuleMatch_SubstringResource(t *testing.T) {
	r := Rule{OperationPattern: "bash:execute", ResourcePattern: "rm -rf", Decision: DecisionDeny}
	if !r.Match("bash:execute", "rm -rf /home") {
		t.Error("expected substring match on resource")
	}
	if !r.Match("bash:execute", "sudo rm -rf /") {
		t.Error("expected substring match anywhere in resource")
	}
	if r.Match("bash:execute", "ls -la") {
		t.Error("should not match when substring not present")
	}
}

func TestRuleMatch_NoMatch(t *testing.T) {
	r := Rule{OperationPattern: "file:write", ResourcePattern: "/etc", Decision: DecisionDeny}
	if r.Match("bash:execute", "/etc/passwd") {
		t.Error("should not match different operation")
	}
	if r.Match("file:write", "/tmp/safe") {
		t.Error("should not match when resource substring not present")
	}
}

// --- StaticPolicy tests ---

func TestStaticPolicy_FirstMatchWins(t *testing.T) {
	p := NewStaticPolicy(DecisionAllow,
		Rule{OperationPattern: "bash:*", ResourcePattern: "rm", Decision: DecisionDeny},
		Rule{OperationPattern: "bash:*", Decision: DecisionAllow},
	)
	if d := p.Evaluate(context.Background(), "bash:execute", "rm -rf /"); d != DecisionDeny {
		t.Errorf("first rule should match and deny, got %v", d)
	}
	if d := p.Evaluate(context.Background(), "bash:execute", "ls"); d != DecisionAllow {
		t.Errorf("second rule should match and allow, got %v", d)
	}
}

func TestStaticPolicy_DefaultDecision(t *testing.T) {
	p := NewStaticPolicy(DecisionDeny)
	if d := p.Evaluate(context.Background(), "anything", "any"); d != DecisionDeny {
		t.Errorf("expected default deny, got %v", d)
	}
}

func TestStaticPolicy_EmptyRules(t *testing.T) {
	p := NewStaticPolicy(DecisionAllow)
	if d := p.Evaluate(context.Background(), "bash:execute", "ls"); d != DecisionAllow {
		t.Errorf("empty rules should use default, got %v", d)
	}
}

func TestStaticPolicy_MultipleRules(t *testing.T) {
	p := NewStaticPolicy(DecisionDeny,
		Rule{OperationPattern: "file:read", Decision: DecisionAllow},
		Rule{OperationPattern: "bash:*", Decision: DecisionAllow},
	)
	if d := p.Evaluate(context.Background(), "file:read", "/tmp/x"); d != DecisionAllow {
		t.Errorf("file:read should be allowed, got %v", d)
	}
	if d := p.Evaluate(context.Background(), "bash:execute", "ls"); d != DecisionAllow {
		t.Errorf("bash should be allowed, got %v", d)
	}
	if d := p.Evaluate(context.Background(), "file:write", "/tmp/x"); d != DecisionDeny {
		t.Errorf("unmatched operation should hit default deny, got %v", d)
	}
}

// --- DefaultPolicy tests ---

func TestDefaultPolicy_AllowsAll(t *testing.T) {
	p := DefaultPolicy()
	ops := []struct {
		op, res string
	}{
		{"bash:execute", "rm -rf /"},
		{"file:write", "/etc/passwd"},
		{"file:read", "/etc/shadow"},
		{"file:edit", "/root/.ssh/authorized_keys"},
		{"grep:execute", "pattern"},
		{"glob:execute", "*.go"},
	}
	for _, tt := range ops {
		if d := p.Evaluate(context.Background(), tt.op, tt.res); d != DecisionAllow {
			t.Errorf("DefaultPolicy(%q, %q) = %v, want allow", tt.op, tt.res, d)
		}
	}
}

// --- SafePolicy tests ---

func TestSafePolicy_DenyDestructiveBash(t *testing.T) {
	p := SafePolicy()
	destructive := []string{
		"rm -rf /",
		"rm -rf /home",
		"sudo rm -rf /",
		"rm -r /*",
		"mkfs.ext4 /dev/sda1",
		"dd if=/dev/zero of=/dev/sda",
		":(){ :|:& };:",
	}
	for _, cmd := range destructive {
		if d := p.Evaluate(context.Background(), "bash:execute", cmd); d != DecisionDeny {
			t.Errorf("SafePolicy should deny %q, got %v", cmd, d)
		}
	}
}

func TestSafePolicy_AllowNonDestructiveBash(t *testing.T) {
	p := SafePolicy()
	safe := []string{"ls -la", "cat file.txt", "go test ./...", "echo hello", "rm old_backup.tar"}
	for _, cmd := range safe {
		if d := p.Evaluate(context.Background(), "bash:execute", cmd); d != DecisionAllow {
			t.Errorf("SafePolicy should allow %q, got %v", cmd, d)
		}
	}
}

func TestSafePolicy_DenyWriteOutsideCWD(t *testing.T) {
	p := SafePolicy()
	cwd, _ := os.Getwd()
	outsidePath := filepath.Join("/tmp", "outside_test_file.txt")
	if d := p.Evaluate(context.Background(), "file:write", outsidePath); d != DecisionDeny {
		t.Errorf("SafePolicy should deny write to %q (cwd=%q), got %v", outsidePath, cwd, d)
	}
	if d := p.Evaluate(context.Background(), "file:edit", outsidePath); d != DecisionDeny {
		t.Errorf("SafePolicy should deny edit to %q (cwd=%q), got %v", outsidePath, cwd, d)
	}
}

func TestSafePolicy_AllowWriteInsideCWD(t *testing.T) {
	p := SafePolicy()
	cwd, _ := os.Getwd()
	insidePath := filepath.Join(cwd, "test_file.txt")
	if d := p.Evaluate(context.Background(), "file:write", insidePath); d != DecisionAllow {
		t.Errorf("SafePolicy should allow write to %q, got %v", insidePath, d)
	}
}

func TestSafePolicy_AllowReads(t *testing.T) {
	p := SafePolicy()
	if d := p.Evaluate(context.Background(), "file:read", "/etc/passwd"); d != DecisionAllow {
		t.Errorf("SafePolicy should allow read anywhere, got %v", d)
	}
}

func TestSafePolicy_AllowGrepGlob(t *testing.T) {
	p := SafePolicy()
	if d := p.Evaluate(context.Background(), "grep:execute", "pattern"); d != DecisionAllow {
		t.Errorf("SafePolicy should allow grep, got %v", d)
	}
	if d := p.Evaluate(context.Background(), "glob:execute", "*.go"); d != DecisionAllow {
		t.Errorf("SafePolicy should allow glob, got %v", d)
	}
}

// --- PolicyProvider adapter tests ---

func TestPolicyProvider_Allow(t *testing.T) {
	p := NewPolicyProvider(DefaultPolicy())
	if err := p.Check(context.Background(), "bash:execute", "ls"); err != nil {
		t.Errorf("expected nil for allow, got %v", err)
	}
}

func TestPolicyProvider_Deny(t *testing.T) {
	p := NewPolicyProvider(SafePolicy())
	err := p.Check(context.Background(), "bash:execute", "rm -rf /")
	if err == nil {
		t.Fatal("expected error for deny")
	}
	want := "permission denied: bash:execute on rm -rf /"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

func TestPolicyProvider_NeedConfirmation(t *testing.T) {
	p := NewPolicyProvider(NewStaticPolicy(DecisionAllow,
		Rule{OperationPattern: "bash:*", Decision: DecisionNeedConfirmation},
	))
	err := p.Check(context.Background(), "bash:execute", "ls")
	if err == nil {
		t.Fatal("expected error for need_confirmation")
	}
	want := "confirmation required: bash:execute on ls"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

func TestPolicyProvider_ImplementsInterface(t *testing.T) {
	// Compile-time check already in adapter.go; this runtime test verifies it works.
	var _ interface {
		Check(ctx context.Context, operation string, resource string) error
	} = NewPolicyProvider(DefaultPolicy())
}
