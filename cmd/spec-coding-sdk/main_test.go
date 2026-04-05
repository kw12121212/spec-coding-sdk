package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestMainExitsZero(t *testing.T) {
	root := projectRoot(t)
	cmd := exec.Command(goBin(), "run", "./cmd/spec-coding-sdk")
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("main should exit zero: %v\noutput: %s", err, out)
	}
}

func goBin() string {
	if goroot := os.Getenv("GOROOT"); goroot != "" {
		return goroot + "/bin/go"
	}
	return "go"
}

func projectRoot(t *testing.T) string {
	t.Helper()
	dir := "."
	for range 10 {
		if _, err := os.Stat(dir + "/go.mod"); err == nil {
			abs, err := filepath.Abs(dir)
			if err != nil {
				t.Fatal(err)
			}
			return abs
		}
		dir += "/.."
	}
	t.Fatal("go.mod not found")
	return ""
}
