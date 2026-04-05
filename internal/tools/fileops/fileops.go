// Package fileops provides file read, write, and edit tools implementing the core.Tool interface.
package fileops

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/kw12121212/spec-coding-sdk/internal/core"
)

// ReadInput is the JSON input schema for ReadTool.
type ReadInput struct {
	FilePath string `json:"file_path"`
	Offset   int    `json:"offset,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}

// WriteInput is the JSON input schema for WriteTool.
type WriteInput struct {
	FilePath string `json:"file_path"`
	Content  string `json:"content"`
}

// EditInput is the JSON input schema for EditTool.
type EditInput struct {
	FilePath   string `json:"file_path"`
	OldString  string `json:"old_string"`
	NewString  string `json:"new_string"`
	ReplaceAll bool   `json:"replace_all,omitempty"`
}

// validateAbsPath validates that the given path is non-empty and absolute.
func validateAbsPath(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("fileops: file_path is required")
	}
	if !filepath.IsAbs(filePath) {
		return fmt.Errorf("fileops: file_path must be absolute, got %q", filePath)
	}
	return nil
}

// checkPerms performs an optional permission check. Returns (result, denied, error).
// If denied is true, the caller should return result immediately.
func checkPerms(ctx context.Context, perms core.PermissionProvider, operation, resource string) (core.ToolResult, bool, error) {
	if perms == nil {
		return core.ToolResult{}, false, nil
	}
	if err := perms.Check(ctx, operation, resource); err != nil {
		return core.ToolResult{IsError: true, Output: err.Error()}, true, nil
	}
	return core.ToolResult{}, false, nil
}
