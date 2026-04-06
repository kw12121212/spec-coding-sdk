// Package builtin provides management for supported external tool binaries.
package builtin

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

const defaultGitHubAPIBaseURL = "https://api.github.com"

// Source identifies where the resolved executable came from.
type Source string

const (
	// SourceSystem means the executable was found in the host PATH.
	SourceSystem Source = "system"
	// SourceManaged means the executable was installed into the SDK-managed directory.
	SourceManaged Source = "managed"
)

// Result describes the resolved executable path and its source.
type Result struct {
	Name   string
	Path   string
	Source Source
}

// Manager resolves supported external tools and installs managed binaries when needed.
type Manager struct {
	installDir     string
	httpClient     *http.Client
	githubAPIBase  string
	supportedTools map[string]toolSpec
	lookPath       func(string) (string, error)
	goos           string
	goarch         string
}

// Option configures a Manager.
type Option func(*Manager)

// WithInstallDir overrides the managed installation root directory.
func WithInstallDir(dir string) Option {
	return func(m *Manager) {
		m.installDir = dir
	}
}

// WithHTTPClient overrides the HTTP client used for release metadata and downloads.
func WithHTTPClient(client *http.Client) Option {
	return func(m *Manager) {
		if client != nil {
			m.httpClient = client
		}
	}
}

// WithGitHubAPIBaseURL overrides the GitHub API base URL.
func WithGitHubAPIBaseURL(baseURL string) Option {
	return func(m *Manager) {
		if baseURL != "" {
			m.githubAPIBase = strings.TrimRight(baseURL, "/")
		}
	}
}

// WithLookPath overrides executable lookup for tests.
func WithLookPath(fn func(string) (string, error)) Option {
	return func(m *Manager) {
		if fn != nil {
			m.lookPath = fn
		}
	}
}

// WithPlatform overrides the platform used for asset selection.
func WithPlatform(goos, goarch string) Option {
	return func(m *Manager) {
		if goos != "" {
			m.goos = goos
		}
		if goarch != "" {
			m.goarch = goarch
		}
	}
}

// NewManager creates a Manager with defaults for rg, gh, and rtk.
func NewManager(opts ...Option) (*Manager, error) {
	installDir, err := defaultInstallDir()
	if err != nil {
		return nil, err
	}

	m := &Manager{
		installDir:     installDir,
		httpClient:     http.DefaultClient,
		githubAPIBase:  defaultGitHubAPIBaseURL,
		supportedTools: defaultToolSpecs(),
		lookPath:       exec.LookPath,
		goos:           runtime.GOOS,
		goarch:         runtime.GOARCH,
	}

	for _, opt := range opts {
		opt(m)
	}

	if m.installDir == "" {
		return nil, fmt.Errorf("builtin: install dir is required")
	}

	return m, nil
}

// Resolve returns an executable path for a supported tool, preferring host-installed binaries.
func (m *Manager) Resolve(ctx context.Context, name string) (Result, error) {
	spec, ok := m.supportedTools[name]
	if !ok {
		return Result{}, fmt.Errorf("builtin: unsupported tool %q", name)
	}

	if path, err := m.lookPath(spec.binaryName); err == nil {
		return Result{Name: name, Path: path, Source: SourceSystem}, nil
	}

	targetPath := m.installTargetPath(spec)
	if info, err := os.Stat(targetPath); err == nil && !info.IsDir() {
		return Result{Name: name, Path: targetPath, Source: SourceManaged}, nil
	}

	assetPattern, ok := spec.assetPattern(m.goos, m.goarch)
	if !ok {
		return Result{}, fmt.Errorf("builtin: tool %q is unsupported on %s/%s", name, m.goos, m.goarch)
	}

	asset, err := m.fetchLatestReleaseAsset(ctx, spec, assetPattern)
	if err != nil {
		return Result{}, err
	}

	if err := m.installAsset(ctx, spec, asset); err != nil {
		return Result{}, err
	}

	return Result{Name: name, Path: targetPath, Source: SourceManaged}, nil
}

type toolSpec struct {
	name         string
	binaryName   string
	owner        string
	repo         string
	assetPattern func(goos, goarch string) (*regexp.Regexp, bool)
}

type releaseAsset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

type githubRelease struct {
	Assets []releaseAsset `json:"assets"`
}

func defaultToolSpecs() map[string]toolSpec {
	return map[string]toolSpec{
		"rg": {
			name:         "rg",
			binaryName:   "rg",
			owner:        "BurntSushi",
			repo:         "ripgrep",
			assetPattern: ripgrepAssetPattern,
		},
		"gh": {
			name:         "gh",
			binaryName:   "gh",
			owner:        "cli",
			repo:         "cli",
			assetPattern: ghAssetPattern,
		},
		"rtk": {
			name:         "rtk",
			binaryName:   "rtk",
			owner:        "rtk-ai",
			repo:         "rtk",
			assetPattern: rtkAssetPattern,
		},
	}
}

func ripgrepAssetPattern(goos, goarch string) (*regexp.Regexp, bool) {
	switch {
	case goos == "linux" && goarch == "amd64":
		return regexp.MustCompile(`^ripgrep-.*-x86_64-unknown-linux-(gnu|musl)\.tar\.gz$`), true
	case goos == "linux" && goarch == "arm64":
		return regexp.MustCompile(`^ripgrep-.*-aarch64-unknown-linux-(gnu|musl)\.tar\.gz$`), true
	case goos == "darwin" && goarch == "amd64":
		return regexp.MustCompile(`^ripgrep-.*-x86_64-apple-darwin\.tar\.gz$`), true
	case goos == "darwin" && goarch == "arm64":
		return regexp.MustCompile(`^ripgrep-.*-aarch64-apple-darwin\.tar\.gz$`), true
	case goos == "windows" && goarch == "amd64":
		return regexp.MustCompile(`^ripgrep-.*-x86_64-pc-windows-msvc\.zip$`), true
	default:
		return nil, false
	}
}

func ghAssetPattern(goos, goarch string) (*regexp.Regexp, bool) {
	switch {
	case goos == "linux" && goarch == "amd64":
		return regexp.MustCompile(`^gh_.*_linux_amd64\.(tar\.gz|zip)$`), true
	case goos == "linux" && goarch == "arm64":
		return regexp.MustCompile(`^gh_.*_linux_arm64\.(tar\.gz|zip)$`), true
	case goos == "darwin" && goarch == "amd64":
		return regexp.MustCompile(`^gh_.*_(macOS|darwin)_amd64\.(tar\.gz|zip)$`), true
	case goos == "darwin" && goarch == "arm64":
		return regexp.MustCompile(`^gh_.*_(macOS|darwin)_arm64\.(tar\.gz|zip)$`), true
	case goos == "windows" && goarch == "amd64":
		return regexp.MustCompile(`^gh_.*_windows_amd64\.zip$`), true
	default:
		return nil, false
	}
}

func rtkAssetPattern(goos, goarch string) (*regexp.Regexp, bool) {
	switch {
	case goos == "linux" && goarch == "amd64":
		return regexp.MustCompile(`^rtk-x86_64-unknown-linux-musl\.tar\.gz$`), true
	case goos == "linux" && goarch == "arm64":
		return regexp.MustCompile(`^rtk-aarch64-unknown-linux-gnu\.tar\.gz$`), true
	case goos == "darwin" && goarch == "amd64":
		return regexp.MustCompile(`^rtk-x86_64-apple-darwin\.tar\.gz$`), true
	case goos == "darwin" && goarch == "arm64":
		return regexp.MustCompile(`^rtk-aarch64-apple-darwin\.tar\.gz$`), true
	case goos == "windows" && goarch == "amd64":
		return regexp.MustCompile(`^rtk-x86_64-pc-windows-msvc\.zip$`), true
	default:
		return nil, false
	}
}

func defaultInstallDir() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("builtin: determine user cache dir: %w", err)
	}
	return filepath.Join(cacheDir, "spec-coding-sdk", "tools"), nil
}

func (m *Manager) installTargetPath(spec toolSpec) string {
	return filepath.Join(m.installDir, spec.name, m.goos+"-"+m.goarch, executableName(spec.binaryName, m.goos))
}

func executableName(name, goos string) string {
	if goos == "windows" {
		return name + ".exe"
	}
	return name
}

func (m *Manager) fetchLatestReleaseAsset(ctx context.Context, spec toolSpec, pattern *regexp.Regexp) (releaseAsset, error) {
	releaseURL := fmt.Sprintf("%s/repos/%s/%s/releases/latest", m.githubAPIBase, spec.owner, spec.repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, releaseURL, nil)
	if err != nil {
		return releaseAsset{}, fmt.Errorf("builtin: create release request for %q: %w", spec.name, err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return releaseAsset{}, fmt.Errorf("builtin: fetch release metadata for %q: %w", spec.name, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return releaseAsset{}, fmt.Errorf("builtin: fetch release metadata for %q: unexpected status %s", spec.name, resp.Status)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return releaseAsset{}, fmt.Errorf("builtin: decode release metadata for %q: %w", spec.name, err)
	}

	for _, asset := range release.Assets {
		if pattern.MatchString(asset.Name) {
			return asset, nil
		}
	}

	return releaseAsset{}, fmt.Errorf("builtin: no release asset found for %q on %s/%s", spec.name, m.goos, m.goarch)
}

func (m *Manager) installAsset(ctx context.Context, spec toolSpec, asset releaseAsset) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, asset.URL, nil)
	if err != nil {
		return fmt.Errorf("builtin: create download request for %q: %w", spec.name, err)
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("builtin: download asset for %q: %w", spec.name, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("builtin: download asset for %q: unexpected status %s", spec.name, resp.Status)
	}

	targetPath := m.installTargetPath(spec)
	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("builtin: create install dir for %q: %w", spec.name, err)
	}

	tmpDir, err := os.MkdirTemp(targetDir, "install-*")
	if err != nil {
		return fmt.Errorf("builtin: create temp install dir for %q: %w", spec.name, err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	tmpFile, err := os.CreateTemp(tmpDir, executableName(spec.binaryName, m.goos)+".*")
	if err != nil {
		return fmt.Errorf("builtin: create temp executable for %q: %w", spec.name, err)
	}
	tmpPath := tmpFile.Name()
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("builtin: close temp executable for %q: %w", spec.name, err)
	}

	expectedName := executableName(spec.binaryName, m.goos)
	if err := extractExecutable(resp.Body, asset.Name, expectedName, tmpPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("builtin: install %q: %w", spec.name, err)
	}

	if m.goos != "windows" {
		if err := os.Chmod(tmpPath, 0o755); err != nil {
			_ = os.Remove(tmpPath)
			return fmt.Errorf("builtin: chmod executable for %q: %w", spec.name, err)
		}
	}

	if err := os.Rename(tmpPath, targetPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("builtin: finalize install for %q: %w", spec.name, err)
	}

	return nil
}

func extractExecutable(src io.Reader, assetName, expectedName, targetPath string) error {
	switch {
	case strings.HasSuffix(assetName, ".tar.gz"):
		return extractFromTarGz(src, expectedName, targetPath)
	case strings.HasSuffix(assetName, ".zip"):
		return extractFromZip(src, expectedName, targetPath)
	default:
		return writeExecutableFile(targetPath, src)
	}
}

func extractFromTarGz(src io.Reader, expectedName, targetPath string) error {
	gzReader, err := gzip.NewReader(src)
	if err != nil {
		return fmt.Errorf("open gzip archive: %w", err)
	}
	defer func() { _ = gzReader.Close() }()

	tarReader := tar.NewReader(gzReader)
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			return fmt.Errorf("expected executable %q not found in archive", expectedName)
		}
		if err != nil {
			return fmt.Errorf("read tar archive: %w", err)
		}
		if header.Typeflag != tar.TypeReg {
			continue
		}
		if filepath.Base(header.Name) != expectedName {
			continue
		}
		return writeExecutableFile(targetPath, tarReader)
	}
}

func extractFromZip(src io.Reader, expectedName, targetPath string) error {
	data, err := io.ReadAll(src)
	if err != nil {
		return fmt.Errorf("read zip archive: %w", err)
	}

	readerAt := bytes.NewReader(data)
	zipReader, err := zip.NewReader(readerAt, int64(len(data)))
	if err != nil {
		return fmt.Errorf("open zip archive: %w", err)
	}

	for _, file := range zipReader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		if filepath.Base(file.Name) != expectedName {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			return fmt.Errorf("open zipped executable: %w", err)
		}
		defer func() { _ = rc.Close() }()
		return writeExecutableFile(targetPath, rc)
	}

	return fmt.Errorf("expected executable %q not found in archive", expectedName)
}

func writeExecutableFile(path string, src io.Reader) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return fmt.Errorf("create executable file: %w", err)
	}
	defer func() { _ = file.Close() }()

	if _, err := io.Copy(file, src); err != nil {
		return fmt.Errorf("write executable file: %w", err)
	}
	return nil
}
