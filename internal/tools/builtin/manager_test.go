package builtin

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
)

func TestResolve_UsesSystemExecutable(t *testing.T) {
	manager := mustManager(t,
		WithInstallDir(t.TempDir()),
		WithLookPath(func(name string) (string, error) {
			if name != "rg" {
				t.Fatalf("expected rg lookup, got %q", name)
			}
			return "/usr/bin/rg", nil
		}),
		WithHTTPClient(rejectingHTTPClient(t)),
	)

	result, err := manager.Resolve(context.Background(), "rg")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if result.Source != SourceSystem {
		t.Fatalf("expected system source, got %q", result.Source)
	}
	if result.Path != "/usr/bin/rg" {
		t.Fatalf("expected system path, got %q", result.Path)
	}
}

func TestResolve_InstallsManagedBinaryFromTarGz(t *testing.T) {
	server := newReleaseServer(t, "rg", runtime.GOOS, runtime.GOARCH, 0)
	defer server.Close()

	manager := mustManager(t,
		WithInstallDir(t.TempDir()),
		WithGitHubAPIBaseURL(server.URL),
		WithHTTPClient(server.Client()),
		WithLookPath(func(string) (string, error) { return "", execNotFound() }),
	)

	result, err := manager.Resolve(context.Background(), "rg")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if result.Source != SourceManaged {
		t.Fatalf("expected managed source, got %q", result.Source)
	}
	content, err := os.ReadFile(result.Path)
	if err != nil {
		t.Fatalf("read installed binary: %v", err)
	}
	if string(content) != "rg-binary" {
		t.Fatalf("unexpected installed content: %q", string(content))
	}
}

func TestResolve_ReusesExistingManagedInstallation(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		calls.Add(1)
		t.Fatal("expected no network calls when managed install exists")
	}))
	defer server.Close()

	manager := mustManager(t,
		WithInstallDir(t.TempDir()),
		WithGitHubAPIBaseURL(server.URL),
		WithHTTPClient(server.Client()),
		WithLookPath(func(string) (string, error) { return "", execNotFound() }),
	)

	target := manager.installTargetPath(manager.supportedTools["gh"])
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(target, []byte("managed-gh"), 0o755); err != nil {
		t.Fatalf("write managed binary: %v", err)
	}

	result, err := manager.Resolve(context.Background(), "gh")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if result.Source != SourceManaged {
		t.Fatalf("expected managed source, got %q", result.Source)
	}
	if result.Path != target {
		t.Fatalf("expected reused path %q, got %q", target, result.Path)
	}
	if calls.Load() != 0 {
		t.Fatalf("expected no network calls, got %d", calls.Load())
	}
}

func TestResolve_UnsupportedTool(t *testing.T) {
	manager := mustManager(t, WithInstallDir(t.TempDir()))

	_, err := manager.Resolve(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected unsupported tool error")
	}
}

func TestResolve_UnsupportedPlatform(t *testing.T) {
	manager := mustManager(t,
		WithInstallDir(t.TempDir()),
		WithPlatform("solaris", "sparc64"),
		WithLookPath(func(string) (string, error) { return "", execNotFound() }),
	)

	_, err := manager.Resolve(context.Background(), "rg")
	if err == nil {
		t.Fatal("expected unsupported platform error")
	}
}

func TestResolve_FailedInstallDoesNotLeaveExecutable(t *testing.T) {
	server := newReleaseServer(t, "rtk", runtime.GOOS, runtime.GOARCH, http.StatusInternalServerError)
	defer server.Close()

	manager := mustManager(t,
		WithInstallDir(t.TempDir()),
		WithGitHubAPIBaseURL(server.URL),
		WithHTTPClient(server.Client()),
		WithLookPath(func(string) (string, error) { return "", execNotFound() }),
	)

	_, err := manager.Resolve(context.Background(), "rtk")
	if err == nil {
		t.Fatal("expected install failure")
	}

	target := manager.installTargetPath(manager.supportedTools["rtk"])
	if _, statErr := os.Stat(target); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("expected no installed executable, got stat err %v", statErr)
	}
}

func mustManager(t *testing.T, opts ...Option) *Manager {
	t.Helper()
	manager, err := NewManager(opts...)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	return manager
}

func rejectingHTTPClient(t *testing.T) *http.Client {
	t.Helper()
	return &http.Client{
		Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			t.Fatal("unexpected HTTP request")
			return nil, nil
		}),
	}
}

func execNotFound() error {
	return errors.New("executable file not found")
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func newReleaseServer(t *testing.T, toolName, goos, goarch string, downloadStatus int) *httptest.Server {
	t.Helper()

	spec := defaultToolSpecs()[toolName]
	pattern, ok := spec.assetPattern(goos, goarch)
	if !ok {
		t.Fatalf("unsupported test platform %s/%s for %s", goos, goarch, toolName)
	}

	assetName := sampleAssetName(t, pattern)
	assetBody := archiveForTool(t, executableName(spec.binaryName, goos), assetName)

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/" + spec.owner + "/" + spec.repo + "/releases/latest":
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(map[string]any{
				"assets": []map[string]string{{
					"name":                 assetName,
					"browser_download_url": "http://" + r.Host + "/download/" + assetName,
				}},
			}); err != nil {
				t.Fatalf("encode release response: %v", err)
			}
		case "/download/" + assetName:
			if downloadStatus != 0 {
				w.WriteHeader(downloadStatus)
				return
			}
			if _, err := w.Write(assetBody); err != nil {
				t.Fatalf("write asset body: %v", err)
			}
		default:
			http.NotFound(w, r)
		}
	}))
}

func sampleAssetName(t *testing.T, pattern *regexp.Regexp) string {
	t.Helper()

	candidates := []string{
		"ripgrep-14.1.1-x86_64-unknown-linux-musl.tar.gz",
		"ripgrep-14.1.1-aarch64-unknown-linux-gnu.tar.gz",
		"ripgrep-14.1.1-x86_64-apple-darwin.tar.gz",
		"ripgrep-14.1.1-aarch64-apple-darwin.tar.gz",
		"ripgrep-14.1.1-x86_64-pc-windows-msvc.zip",
		"gh_2.83.0_linux_amd64.tar.gz",
		"gh_2.83.0_linux_arm64.tar.gz",
		"gh_2.83.0_macOS_amd64.zip",
		"gh_2.83.0_macOS_arm64.zip",
		"gh_2.83.0_windows_amd64.zip",
		"rtk-x86_64-unknown-linux-musl.tar.gz",
		"rtk-aarch64-unknown-linux-gnu.tar.gz",
		"rtk-x86_64-apple-darwin.tar.gz",
		"rtk-aarch64-apple-darwin.tar.gz",
		"rtk-x86_64-pc-windows-msvc.zip",
	}

	for _, candidate := range candidates {
		if pattern.MatchString(candidate) {
			return candidate
		}
	}

	t.Fatalf("no sample asset matched pattern %q", pattern.String())
	return ""
}

func archiveForTool(t *testing.T, executable, assetName string) []byte {
	t.Helper()

	content := []byte(strings.TrimSuffix(executable, ".exe") + "-binary")
	if strings.HasSuffix(assetName, ".zip") {
		var buf bytes.Buffer
		zipWriter := zip.NewWriter(&buf)
		file, err := zipWriter.Create("bundle/" + executable)
		if err != nil {
			t.Fatalf("create zip entry: %v", err)
		}
		if _, err := file.Write(content); err != nil {
			t.Fatalf("write zip entry: %v", err)
		}
		if err := zipWriter.Close(); err != nil {
			t.Fatalf("close zip writer: %v", err)
		}
		return buf.Bytes()
	}

	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	tarWriter := tar.NewWriter(gzWriter)
	header := &tar.Header{
		Name: "bundle/" + executable,
		Mode: 0o755,
		Size: int64(len(content)),
	}
	if err := tarWriter.WriteHeader(header); err != nil {
		t.Fatalf("write tar header: %v", err)
	}
	if _, err := tarWriter.Write(content); err != nil {
		t.Fatalf("write tar entry: %v", err)
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatalf("close tar writer: %v", err)
	}
	if err := gzWriter.Close(); err != nil {
		t.Fatalf("close gzip writer: %v", err)
	}
	return buf.Bytes()
}
