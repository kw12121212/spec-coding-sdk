// Package gh provides a GitHub CLI tool implementing the core.Tool interface.
package gh

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/kw12121212/spec-coding-sdk/internal/core"
	"github.com/kw12121212/spec-coding-sdk/internal/tools/builtin"
)

const (
	defaultTimeoutSec = 120
	maxOutputBytes    = 1 << 20 // 1MB
	truncationNotice  = "\n...[output truncated: exceeded 1MB limit]"
)

// Input is the JSON input schema for Tool.
type Input struct {
	Args       []string `json:"args"`
	WorkingDir string   `json:"working_dir,omitempty"`
	Timeout    int      `json:"timeout,omitempty"`
}

type resolvedExecutable struct {
	Path string
}

type executableResolver interface {
	Resolve(ctx context.Context, name string) (resolvedExecutable, error)
}

type defaultResolver struct{}

func (defaultResolver) Resolve(ctx context.Context, name string) (resolvedExecutable, error) {
	manager, err := builtin.NewManager()
	if err != nil {
		return resolvedExecutable{}, fmt.Errorf("gh: create builtin manager: %w", err)
	}

	result, err := manager.Resolve(ctx, name)
	if err != nil {
		return resolvedExecutable{}, fmt.Errorf("gh: resolve executable: %w", err)
	}

	return resolvedExecutable{Path: result.Path}, nil
}

// Tool executes GitHub CLI commands.
type Tool struct {
	perms       core.PermissionProvider
	resolver    executableResolver
	execCommand func(context.Context, string, ...string) *exec.Cmd
}

// New creates a new Tool. A nil perms means no permission checks.
func New(perms core.PermissionProvider) *Tool {
	return &Tool{
		perms:    perms,
		resolver: defaultResolver{},
	}
}

// Execute implements core.Tool.
func (t *Tool) Execute(ctx context.Context, input json.RawMessage) (core.ToolResult, error) {
	var in Input
	if err := json.Unmarshal(input, &in); err != nil {
		return core.ToolResult{}, fmt.Errorf("gh: invalid input JSON: %w", err)
	}
	if len(in.Args) == 0 {
		return core.ToolResult{}, fmt.Errorf("gh: args are required")
	}
	for _, arg := range in.Args {
		if arg == "" {
			return core.ToolResult{}, fmt.Errorf("gh: args must not contain empty values")
		}
	}

	if t.perms != nil {
		if err := t.perms.Check(ctx, "gh:execute", strings.Join(in.Args, " ")); err != nil {
			return core.ToolResult{IsError: true, Output: err.Error()}, nil
		}
	}

	timeoutSec := in.Timeout
	if timeoutSec <= 0 {
		timeoutSec = defaultTimeoutSec
	}

	execCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
	defer cancel()

	executable, err := t.getResolver().Resolve(execCtx, "gh")
	if err != nil {
		return core.ToolResult{IsError: true, Output: err.Error()}, nil
	}

	cmd := t.commandExecutor()(execCtx, executable.Path, in.Args...)
	if in.WorkingDir != "" {
		cmd.Dir = in.WorkingDir
	}

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	runErr := cmd.Run()
	if execCtx.Err() == context.DeadlineExceeded {
		return core.ToolResult{
			IsError: true,
			Output:  "gh: command timed out",
		}, nil
	}

	output, truncated := truncateOutput(buf.String())
	if runErr != nil {
		if output == "" {
			output = runErr.Error()
		}
		return core.ToolResult{
			IsError: true,
			Output:  output,
		}, nil
	}

	return core.ToolResult{
		IsError: truncated,
		Output:  output,
	}, nil
}

func (t *Tool) getResolver() executableResolver {
	if t.resolver != nil {
		return t.resolver
	}

	return defaultResolver{}
}

func (t *Tool) commandExecutor() func(context.Context, string, ...string) *exec.Cmd {
	if t.execCommand != nil {
		return t.execCommand
	}

	return exec.CommandContext
}

func truncateOutput(output string) (string, bool) {
	if len(output) <= maxOutputBytes {
		return output, false
	}

	return output[:maxOutputBytes] + truncationNotice, true
}
