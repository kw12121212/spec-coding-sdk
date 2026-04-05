// Package glob provides a file pattern matching tool implementing the core.Tool interface.
package glob

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kw12121212/spec-coding-sdk/internal/core"
)

const (
	maxOutputBytes   = 1 << 20 // 1MB
	truncationNotice = "\n...[output truncated: exceeded 1MB limit]"
)

// Input is the JSON input schema for Tool.
type Input struct {
	Pattern string `json:"pattern"`
	Path    string `json:"path,omitempty"`
}

// Tool finds files by glob pattern, supporting recursive ** matching.
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
		return core.ToolResult{}, fmt.Errorf("glob: invalid input JSON: %w", err)
	}
	if in.Pattern == "" {
		return core.ToolResult{}, fmt.Errorf("glob: pattern is required")
	}

	// Permission check
	if t.perms != nil {
		if err := t.perms.Check(ctx, "glob:execute", in.Pattern); err != nil {
			return core.ToolResult{IsError: true, Output: err.Error()}, nil
		}
	}

	baseDir := in.Path
	if baseDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return core.ToolResult{}, fmt.Errorf("glob: failed to get working directory: %w", err)
		}
		baseDir = wd
	}

	info, err := os.Stat(baseDir)
	if err != nil {
		return core.ToolResult{IsError: true, Output: fmt.Sprintf("glob: path %q does not exist", baseDir)}, nil
	}
	if !info.IsDir() {
		return core.ToolResult{IsError: true, Output: fmt.Sprintf("glob: path %q is not a directory", baseDir)}, nil
	}

	var matches []fileMatch
	err = filepath.WalkDir(baseDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(baseDir, path)
		if err != nil {
			return nil
		}
		if matchGlob(rel, in.Pattern) {
			info, err := d.Info()
			modTime := time.Time{}
			if err == nil {
				modTime = info.ModTime()
			}
			matches = append(matches, fileMatch{path: path, modTime: modTime})
		}
		return nil
	})
	if err != nil {
		return core.ToolResult{IsError: true, Output: fmt.Sprintf("glob: walk error: %s", err)}, nil
	}

	// Sort by modification time, most recent first
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].modTime.After(matches[j].modTime)
	})

	var lines []string
	for _, m := range matches {
		lines = append(lines, m.path)
	}
	output := strings.Join(lines, "\n")

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

type fileMatch struct {
	path    string
	modTime time.Time
}

// matchGlob matches a relative path against a glob pattern supporting *, ?, and **.
func matchGlob(relPath, pattern string) bool {
	// Normalize separators to /
	relPath = filepath.ToSlash(relPath)
	pattern = filepath.ToSlash(pattern)

	// Fast path: no ** involved
	if !strings.Contains(pattern, "**") {
		matched, _ := filepath.Match(pattern, relPath)
		return matched
	}

	return matchDoublestar(strings.Split(relPath, "/"), strings.Split(pattern, "/"))
}

// matchDoublestar performs segment-by-segment matching with ** support.
func matchDoublestar(pathParts, patternParts []string) bool {
	for len(patternParts) > 0 {
		seg := patternParts[0]
		if seg == "**" {
			patternParts = patternParts[1:]
			// ** at end matches everything remaining
			if len(patternParts) == 0 {
				return true
			}
			// Try matching remaining pattern at every position
			for i := 0; i <= len(pathParts); i++ {
				if matchDoublestar(pathParts[i:], patternParts) {
					return true
				}
			}
			return false
		}

		if len(pathParts) == 0 {
			return false
		}
		matched, _ := filepath.Match(seg, pathParts[0])
		if !matched {
			return false
		}
		pathParts = pathParts[1:]
		patternParts = patternParts[1:]
	}

	return len(pathParts) == 0
}
