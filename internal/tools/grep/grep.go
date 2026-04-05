// Package grep provides a content search tool based on ripgrep, implementing the core.Tool interface.
package grep

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/kw12121212/spec-coding-sdk/internal/core"
)

const (
	maxOutputBytes   = 1 << 20 // 1MB
	truncationNotice = "\n...[output truncated: exceeded 1MB limit]"
)

// Input is the JSON input schema for Tool.
type Input struct {
	Pattern     string `json:"pattern"`
	Path        string `json:"path,omitempty"`
	Glob        string `json:"glob,omitempty"`
	Type        string `json:"type,omitempty"`
	OutputMode  string `json:"output_mode,omitempty"`
	IgnoreCase  bool   `json:"ignore_case,omitempty"`
	Context     int    `json:"context,omitempty"`
	HeadLimit   int    `json:"head_limit,omitempty"`
}

// Tool searches file contents using ripgrep.
type Tool struct {
	perms core.PermissionProvider
}

// New creates a new Tool. A nil perms means no permission checks.
func New(perms core.PermissionProvider) *Tool {
	return &Tool{perms: perms}
}

// Execute implements core.Tool.
func (t *Tool) Execute(ctx context.Context, input json.RawMessage) (core.ToolResult, error) {
	var in Input
	if err := json.Unmarshal(input, &in); err != nil {
		return core.ToolResult{}, fmt.Errorf("grep: invalid input JSON: %w", err)
	}
	if in.Pattern == "" {
		return core.ToolResult{}, fmt.Errorf("grep: pattern is required")
	}

	// Permission check
	if t.perms != nil {
		if err := t.perms.Check(ctx, "grep:execute", in.Pattern); err != nil {
			return core.ToolResult{IsError: true, Output: err.Error()}, nil
		}
	}

	args := buildArgs(in)

	cmd := exec.CommandContext(ctx, "rg", args...)
	if in.Path != "" {
		cmd.Args = append(cmd.Args, in.Path)
	}

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	runErr := cmd.Run()
	output := buf.String()

	// rg exit codes: 0 = match found, 1 = no match, 2+ = error
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode := exitErr.ExitCode()
			if exitCode == 1 {
				return core.ToolResult{
					IsError: false,
					Output:  "grep: no matches found",
				}, nil
			}
			if exitCode > 1 {
				return core.ToolResult{
					IsError: true,
					Output:  output,
				}, nil
			}
		}
		// rg binary not found
		if runErr == exec.ErrNotFound || strings.Contains(runErr.Error(), "executable file not found") {
			return core.ToolResult{
				IsError: true,
				Output:  "grep: rg not found in PATH",
			}, nil
		}
		return core.ToolResult{IsError: true, Output: runErr.Error()}, nil
	}

	// Apply head_limit
	if in.HeadLimit > 0 {
		lines := strings.SplitN(output, "\n", in.HeadLimit+1)
		if len(lines) > in.HeadLimit {
			output = strings.Join(lines[:in.HeadLimit], "\n")
		}
	}

	// Truncate output at 1MB
	truncated := false
	if len(output) > maxOutputBytes {
		output = output[:maxOutputBytes] + truncationNotice
		truncated = true
	}

	return core.ToolResult{
		Output:  output,
		IsError: truncated,
	}, nil
}

func buildArgs(in Input) []string {
	args := []string{"--no-heading"}

	switch in.OutputMode {
	case "files_with_matches":
		args = append(args, "-l")
	case "count":
		args = append(args, "-c")
	default:
		// content mode: show line numbers
		args = append(args, "-n")
	}

	if in.IgnoreCase {
		args = append(args, "-i")
	}
	if in.Glob != "" {
		args = append(args, "--glob", in.Glob)
	}
	if in.Type != "" {
		args = append(args, "--type", in.Type)
	}
	if in.Context > 0 {
		args = append(args, "-C", fmt.Sprintf("%d", in.Context))
	}

	args = append(args, in.Pattern)
	return args
}
