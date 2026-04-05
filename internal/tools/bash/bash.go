// Package bash provides a Bash command execution tool implementing the core.Tool interface.
package bash

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/kw12121212/spec-coding-sdk/internal/core"
)

const (
	defaultTimeoutSec = 120
	maxOutputBytes    = 1 << 20 // 1MB
	truncationNotice  = "\n...[output truncated: exceeded 1MB limit]"
)

// Input is the JSON input schema for Tool.
type Input struct {
	Command    string `json:"command"`
	Timeout    int    `json:"timeout,omitempty"`
	WorkingDir string `json:"working_dir,omitempty"`
}

// Tool executes bash commands via /bin/bash -c.
type Tool struct {
	perms core.PermissionProvider
}

// New creates a new Tool. Options can provide a PermissionProvider.
func New(perms core.PermissionProvider) *Tool {
	return &Tool{perms: perms}
}

// Execute implements core.Tool.
func (t *Tool) Execute(ctx context.Context, input json.RawMessage) (core.ToolResult, error) {
	var in Input
	if err := json.Unmarshal(input, &in); err != nil {
		return core.ToolResult{}, fmt.Errorf("bash: invalid input JSON: %w", err)
	}
	if in.Command == "" {
		return core.ToolResult{}, fmt.Errorf("bash: command is required")
	}

	// Permission check
	if t.perms != nil {
		if err := t.perms.Check(ctx, "bash:execute", in.Command); err != nil {
			return core.ToolResult{IsError: true, Output: err.Error()}, nil
		}
	}

	timeoutSec := in.Timeout
	if timeoutSec <= 0 {
		timeoutSec = defaultTimeoutSec
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "/bin/bash", "-c", in.Command)
	if in.WorkingDir != "" {
		cmd.Dir = in.WorkingDir
	}

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	runErr := cmd.Run()
	output := buf.String()

	// Handle timeout
	if ctx.Err() == context.DeadlineExceeded {
		return core.ToolResult{
			IsError: true,
			Output:  "bash: command timed out",
		}, nil
	}

	// Handle output size limit
	truncated := false
	if len(output) > maxOutputBytes {
		output = output[:maxOutputBytes] + truncationNotice
		truncated = true
	}

	isError := truncated || runErr != nil
	return core.ToolResult{Output: output, IsError: isError}, nil
}
