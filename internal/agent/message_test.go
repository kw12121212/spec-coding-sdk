package agent

import (
	"testing"
)

func TestRoleConstants(t *testing.T) {
	if RoleUser != "user" {
		t.Fatalf("expected RoleUser %q, got %q", "user", RoleUser)
	}
	if RoleAssistant != "assistant" {
		t.Fatalf("expected RoleAssistant %q, got %q", "assistant", RoleAssistant)
	}
	if RoleTool != "tool" {
		t.Fatalf("expected RoleTool %q, got %q", "tool", RoleTool)
	}
}

func TestNewMessage(t *testing.T) {
	msg := NewMessage(RoleUser, "hello")
	if msg.Role != RoleUser {
		t.Fatalf("expected role %q, got %q", RoleUser, msg.Role)
	}
	if msg.Content != "hello" {
		t.Fatalf("expected content 'hello', got %q", msg.Content)
	}
	if msg.Timestamp.IsZero() {
		t.Fatal("expected non-zero timestamp")
	}
	if msg.ToolName != "" {
		t.Fatalf("expected empty ToolName, got %q", msg.ToolName)
	}
}

func TestNewToolMessage(t *testing.T) {
	msg := NewToolMessage("bash", "output")
	if msg.Role != RoleTool {
		t.Fatalf("expected role %q, got %q", RoleTool, msg.Role)
	}
	if msg.Content != "output" {
		t.Fatalf("expected content 'output', got %q", msg.Content)
	}
	if msg.ToolName != "bash" {
		t.Fatalf("expected ToolName 'bash', got %q", msg.ToolName)
	}
	if msg.Timestamp.IsZero() {
		t.Fatal("expected non-zero timestamp")
	}
}
