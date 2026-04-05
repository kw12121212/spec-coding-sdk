package fileops

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kw12121212/spec-coding-sdk/internal/core"
)

// WriteTool creates or overwrites files with automatic parent directory creation.
type WriteTool struct {
	perms core.PermissionProvider
}

// NewWriteTool creates a new WriteTool with optional permission provider.
func NewWriteTool(perms core.PermissionProvider) *WriteTool {
	return &WriteTool{perms: perms}
}

// Execute implements core.Tool.
func (t *WriteTool) Execute(ctx context.Context, input json.RawMessage) (core.ToolResult, error) {
	var in WriteInput
	if err := json.Unmarshal(input, &in); err != nil {
		return core.ToolResult{}, fmt.Errorf("fileops write: invalid input JSON: %w", err)
	}

	if err := validateAbsPath(in.FilePath); err != nil {
		return core.ToolResult{}, err
	}

	if result, denied, err := checkPerms(ctx, t.perms, "file:write", in.FilePath); denied || err != nil {
		return result, err
	}

	dir := filepath.Dir(in.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return core.ToolResult{IsError: true, Output: fmt.Sprintf("fileops write: failed to create directory: %v", err)}, nil
	}

	if err := os.WriteFile(in.FilePath, []byte(in.Content), 0644); err != nil {
		return core.ToolResult{IsError: true, Output: fmt.Sprintf("fileops write: %v", err)}, nil
	}

	return core.ToolResult{Output: fmt.Sprintf("fileops write: wrote %s", in.FilePath)}, nil
}
