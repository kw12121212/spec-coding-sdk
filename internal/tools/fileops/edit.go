package fileops

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/kw12121212/spec-coding-sdk/internal/core"
)

// EditTool performs precise string replacements in files.
type EditTool struct {
	perms core.PermissionProvider
}

// NewEditTool creates a new EditTool with optional permission provider.
func NewEditTool(perms core.PermissionProvider) *EditTool {
	return &EditTool{perms: perms}
}

// Execute implements core.Tool.
func (t *EditTool) Execute(ctx context.Context, input json.RawMessage) (core.ToolResult, error) {
	var in EditInput
	if err := json.Unmarshal(input, &in); err != nil {
		return core.ToolResult{}, fmt.Errorf("fileops edit: invalid input JSON: %w", err)
	}

	if err := validateAbsPath(in.FilePath); err != nil {
		return core.ToolResult{}, err
	}

	if in.OldString == "" {
		return core.ToolResult{}, fmt.Errorf("fileops edit: old_string is required")
	}

	if result, denied, err := checkPerms(ctx, t.perms, "file:edit", in.FilePath); denied || err != nil {
		return result, err
	}

	data, err := os.ReadFile(in.FilePath)
	if err != nil {
		return core.ToolResult{IsError: true, Output: fmt.Sprintf("fileops edit: %v", err)}, nil
	}

	content := string(data)
	count := strings.Count(content, in.OldString)

	if count == 0 {
		return core.ToolResult{
			IsError: true,
			Output:  "fileops edit: old_string not found in file",
		}, nil
	}

	if !in.ReplaceAll && count > 1 {
		return core.ToolResult{
			IsError: true,
			Output:  fmt.Sprintf("fileops edit: old_string matches %d locations; must be unique or set replace_all=true", count),
		}, nil
	}

	var newContent string
	if in.ReplaceAll {
		newContent = strings.ReplaceAll(content, in.OldString, in.NewString)
	} else {
		newContent = strings.Replace(content, in.OldString, in.NewString, 1)
	}

	if err := os.WriteFile(in.FilePath, []byte(newContent), 0644); err != nil {
		return core.ToolResult{IsError: true, Output: fmt.Sprintf("fileops edit: %v", err)}, nil
	}

	occurrences := count
	if in.ReplaceAll {
		return core.ToolResult{Output: fmt.Sprintf("fileops edit: replaced %d occurrences in %s", occurrences, in.FilePath)}, nil
	}
	return core.ToolResult{Output: fmt.Sprintf("fileops edit: replaced 1 occurrence in %s", in.FilePath)}, nil
}
