package fileops

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/kw12121212/spec-coding-sdk/internal/core"
)

// ReadTool reads file contents with optional line number formatting and range selection.
type ReadTool struct {
	perms core.PermissionProvider
}

// NewReadTool creates a new ReadTool with optional permission provider.
func NewReadTool(perms core.PermissionProvider) *ReadTool {
	return &ReadTool{perms: perms}
}

// Execute implements core.Tool.
func (t *ReadTool) Execute(ctx context.Context, input json.RawMessage) (core.ToolResult, error) {
	var in ReadInput
	if err := json.Unmarshal(input, &in); err != nil {
		return core.ToolResult{}, fmt.Errorf("fileops read: invalid input JSON: %w", err)
	}

	if err := validateAbsPath(in.FilePath); err != nil {
		return core.ToolResult{}, err
	}

	if result, denied, err := checkPerms(ctx, t.perms, "file:read", in.FilePath); denied || err != nil {
		return result, err
	}

	data, err := os.ReadFile(in.FilePath)
	if err != nil {
		return core.ToolResult{IsError: true, Output: fmt.Sprintf("fileops read: %v", err)}, nil
	}

	lines := strings.Split(string(data), "\n")
	// If the file ends with a newline, remove the trailing empty element.
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	// Apply offset (1-based to 0-based).
	if in.Offset > 0 {
		if in.Offset > len(lines) {
			lines = nil
		} else {
			lines = lines[in.Offset-1:]
		}
	}

	// Apply limit.
	if in.Limit > 0 && in.Limit < len(lines) {
		lines = lines[:in.Limit]
	}

	var b strings.Builder
	for i, line := range lines {
		lineNum := i + 1
		if in.Offset > 0 {
			lineNum = in.Offset + i
		}
		fmt.Fprintf(&b, "%d\t%s\n", lineNum, line)
	}

	return core.ToolResult{Output: b.String()}, nil
}
