package core_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kw12121212/spec-coding-sdk/internal/core"
)

func TestLoadConfigValidEmptyYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(""), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := core.LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if cfg == nil {
		t.Fatal("LoadConfig returned nil Config")
	}
}

func TestLoadConfigFileNotFound(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.yaml")

	_, err := core.LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}

		// Error must contain the file path.
	if !strings.Contains(err.Error(), "nonexistent.yaml") {
		t.Fatalf("error %q does not contain file path", err.Error())
	}
}

func TestLoadConfigMalformedYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(":\n  :\n    - [invalid"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := core.LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for malformed YAML")
	}

	// Error must describe the parse failure.
	if !strings.Contains(err.Error(), "parse") {
		t.Fatalf("error %q does not describe parse failure", err.Error())
	}
}

// TestConfigImportable verifies that Config can be referenced from an external package.
func TestConfigImportable(_ *testing.T) {
	_ = &core.Config{}
}
