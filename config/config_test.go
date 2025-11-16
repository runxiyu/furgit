package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func setupTestRepo(t *testing.T) (string, func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "furgit-config-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	cleanup := func() {
		_ = os.RemoveAll(tempDir)
	}

	cmd := exec.Command("git", "init", "--object-format=sha256", "--bare", tempDir)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
	if output, err := cmd.CombinedOutput(); err != nil {
		cleanup()
		t.Fatalf("failed to init git repo: %v\n%s", err, output)
	}

	return tempDir, cleanup
}

func gitConfig(t *testing.T, dir string, args ...string) {
	t.Helper()
	fullArgs := append([]string{"config"}, args...)
	cmd := exec.Command("git", fullArgs...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git config %v failed: %v\n%s", args, err, output)
	}
}

func gitConfigGet(t *testing.T, dir, key string) string {
	t.Helper()
	cmd := exec.Command("git", "config", "--get", key)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func TestConfigAgainstGit(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	gitConfig(t, repoPath, "core.bare", "true")
	gitConfig(t, repoPath, "core.filemode", "false")
	gitConfig(t, repoPath, "user.name", "John Doe")
	gitConfig(t, repoPath, "user.email", "john@example.com")

	cfgFile, err := os.Open(filepath.Join(repoPath, "config"))
	if err != nil {
		t.Fatalf("failed to open config: %v", err)
	}
	defer func() { _ = cfgFile.Close() }()

	cfg, err := ParseConfig(cfgFile)
	if err != nil {
		t.Fatalf("ParseConfig failed: %v", err)
	}

	if got := cfg.Get("core", "", "bare"); got != "true" {
		t.Errorf("core.bare: got %q, want %q", got, "true")
	}
	if got := cfg.Get("core", "", "filemode"); got != "false" {
		t.Errorf("core.filemode: got %q, want %q", got, "false")
	}
	if got := cfg.Get("user", "", "name"); got != "John Doe" {
		t.Errorf("user.name: got %q, want %q", got, "John Doe")
	}
	if got := cfg.Get("user", "", "email"); got != "john@example.com" {
		t.Errorf("user.email: got %q, want %q", got, "john@example.com")
	}
}

func TestConfigSubsectionAgainstGit(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	gitConfig(t, repoPath, "remote.origin.url", "https://example.com/repo.git")
	gitConfig(t, repoPath, "remote.origin.fetch", "+refs/heads/*:refs/remotes/origin/*")

	cfgFile, err := os.Open(filepath.Join(repoPath, "config"))
	if err != nil {
		t.Fatalf("failed to open config: %v", err)
	}
	defer func() { _ = cfgFile.Close() }()

	cfg, err := ParseConfig(cfgFile)
	if err != nil {
		t.Fatalf("ParseConfig failed: %v", err)
	}

	if got := cfg.Get("remote", "origin", "url"); got != "https://example.com/repo.git" {
		t.Errorf("remote.origin.url: got %q, want %q", got, "https://example.com/repo.git")
	}
	if got := cfg.Get("remote", "origin", "fetch"); got != "+refs/heads/*:refs/remotes/origin/*" {
		t.Errorf("remote.origin.fetch: got %q, want %q", got, "+refs/heads/*:refs/remotes/origin/*")
	}
}

func TestConfigMultiValueAgainstGit(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	gitConfig(t, repoPath, "--add", "remote.origin.fetch", "+refs/heads/main:refs/remotes/origin/main")
	gitConfig(t, repoPath, "--add", "remote.origin.fetch", "+refs/heads/dev:refs/remotes/origin/dev")
	gitConfig(t, repoPath, "--add", "remote.origin.fetch", "+refs/tags/*:refs/tags/*")

	cfgFile, err := os.Open(filepath.Join(repoPath, "config"))
	if err != nil {
		t.Fatalf("failed to open config: %v", err)
	}
	defer func() { _ = cfgFile.Close() }()

	cfg, err := ParseConfig(cfgFile)
	if err != nil {
		t.Fatalf("ParseConfig failed: %v", err)
	}

	fetches := cfg.GetAll("remote", "origin", "fetch")
	if len(fetches) != 3 {
		t.Fatalf("expected 3 fetch values, got %d", len(fetches))
	}

	expected := []string{
		"+refs/heads/main:refs/remotes/origin/main",
		"+refs/heads/dev:refs/remotes/origin/dev",
		"+refs/tags/*:refs/tags/*",
	}
	for i, want := range expected {
		if fetches[i] != want {
			t.Errorf("fetch[%d]: got %q, want %q", i, fetches[i], want)
		}
	}
}

func TestConfigCaseInsensitiveAgainstGit(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	gitConfig(t, repoPath, "Core.Bare", "true")
	gitConfig(t, repoPath, "CORE.FileMode", "false")

	gitVerifyBare := gitConfigGet(t, repoPath, "core.bare")
	gitVerifyFilemode := gitConfigGet(t, repoPath, "core.filemode")

	cfgFile, err := os.Open(filepath.Join(repoPath, "config"))
	if err != nil {
		t.Fatalf("failed to open config: %v", err)
	}
	defer func() { _ = cfgFile.Close() }()

	cfg, err := ParseConfig(cfgFile)
	if err != nil {
		t.Fatalf("ParseConfig failed: %v", err)
	}

	if got := cfg.Get("core", "", "bare"); got != gitVerifyBare {
		t.Errorf("core.bare: got %q, want %q (from git)", got, gitVerifyBare)
	}
	if got := cfg.Get("CORE", "", "BARE"); got != gitVerifyBare {
		t.Errorf("CORE.BARE: got %q, want %q (from git)", got, gitVerifyBare)
	}
	if got := cfg.Get("core", "", "filemode"); got != gitVerifyFilemode {
		t.Errorf("core.filemode: got %q, want %q (from git)", got, gitVerifyFilemode)
	}
}

func TestConfigBooleanAgainstGit(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	gitConfig(t, repoPath, "test.flag1", "true")
	gitConfig(t, repoPath, "test.flag2", "false")
	gitConfig(t, repoPath, "test.flag3", "yes")
	gitConfig(t, repoPath, "test.flag4", "no")

	cfgFile, err := os.Open(filepath.Join(repoPath, "config"))
	if err != nil {
		t.Fatalf("failed to open config: %v", err)
	}
	defer func() { _ = cfgFile.Close() }()

	cfg, err := ParseConfig(cfgFile)
	if err != nil {
		t.Fatalf("ParseConfig failed: %v", err)
	}

	tests := []struct {
		key  string
		want string
	}{
		{"flag1", gitConfigGet(t, repoPath, "test.flag1")},
		{"flag2", gitConfigGet(t, repoPath, "test.flag2")},
		{"flag3", gitConfigGet(t, repoPath, "test.flag3")},
		{"flag4", gitConfigGet(t, repoPath, "test.flag4")},
	}

	for _, tt := range tests {
		if got := cfg.Get("test", "", tt.key); got != tt.want {
			t.Errorf("test.%s: got %q, want %q (from git)", tt.key, got, tt.want)
		}
	}
}

func TestConfigComplexValuesAgainstGit(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	gitConfig(t, repoPath, "test.spaced", "value with spaces")
	gitConfig(t, repoPath, "test.special", "value=with=equals")
	gitConfig(t, repoPath, "test.path", "/path/to/something")
	gitConfig(t, repoPath, "test.number", "12345")

	cfgFile, err := os.Open(filepath.Join(repoPath, "config"))
	if err != nil {
		t.Fatalf("failed to open config: %v", err)
	}
	defer func() { _ = cfgFile.Close() }()

	cfg, err := ParseConfig(cfgFile)
	if err != nil {
		t.Fatalf("ParseConfig failed: %v", err)
	}

	tests := []string{"spaced", "special", "path", "number"}
	for _, key := range tests {
		want := gitConfigGet(t, repoPath, "test."+key)
		if got := cfg.Get("test", "", key); got != want {
			t.Errorf("test.%s: got %q, want %q (from git)", key, got, want)
		}
	}
}

func TestConfigEntriesAgainstGit(t *testing.T) {
	repoPath, cleanup := setupTestRepo(t)
	defer cleanup()

	gitConfig(t, repoPath, "core.bare", "true")
	gitConfig(t, repoPath, "core.filemode", "false")
	gitConfig(t, repoPath, "user.name", "Test User")

	cfgFile, err := os.Open(filepath.Join(repoPath, "config"))
	if err != nil {
		t.Fatalf("failed to open config: %v", err)
	}
	defer func() { _ = cfgFile.Close() }()

	cfg, err := ParseConfig(cfgFile)
	if err != nil {
		t.Fatalf("ParseConfig failed: %v", err)
	}

	entries := cfg.Entries()
	if len(entries) < 3 {
		t.Errorf("expected at least 3 entries, got %d", len(entries))
	}

	found := make(map[string]bool)
	for _, entry := range entries {
		key := entry.Section + "." + entry.Key
		if entry.Subsection != "" {
			key = entry.Section + "." + entry.Subsection + "." + entry.Key
		}
		found[key] = true

		gitValue := gitConfigGet(t, repoPath, key)
		if entry.Value != gitValue {
			t.Errorf("entry %s: got value %q, git has %q", key, entry.Value, gitValue)
		}
	}
}

func TestConfigErrorCases(t *testing.T) {
	tests := []struct {
		name   string
		config string
	}{
		{
			name:   "key before section",
			config: "bare = true",
		},
		{
			name:   "invalid section character",
			config: "[core/invalid]",
		},
		{
			name:   "unterminated section",
			config: "[core",
		},
		{
			name:   "unterminated quote",
			config: "[core]\n\tbare = \"true",
		},
		{
			name:   "invalid escape",
			config: "[core]\n\tvalue = \"test\\x\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.config)
			_, err := ParseConfig(r)
			if err == nil {
				t.Errorf("expected error for %s", tt.name)
			}
		})
	}
}
