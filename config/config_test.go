package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"codeberg.org/lindenii/furgit/config"
	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/objectid"
)

func openConfig(t *testing.T, testRepo *testgit.TestRepo) *os.File {
	t.Helper()
	cfgFile, err := os.Open(filepath.Join(testRepo.Dir(), "config"))
	if err != nil {
		t.Fatalf("failed to open config: %v", err)
	}
	return cfgFile
}

func gitConfigGet(t *testing.T, testRepo *testgit.TestRepo, key string) string {
	t.Helper()
	return testRepo.Run(t, "config", "--get", key)
}

func TestConfigAgainstGit(t *testing.T) {
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		testRepo.Run(t, "config", "core.bare", "true")
		testRepo.Run(t, "config", "core.filemode", "false")
		testRepo.Run(t, "config", "user.name", "Jane Doe")
		testRepo.Run(t, "config", "user.email", "jane@example.org")

		cfgFile := openConfig(t, testRepo)
		defer func() { _ = cfgFile.Close() }()

		cfg, err := config.ParseConfig(cfgFile)
		if err != nil {
			t.Fatalf("ParseConfig failed: %v", err)
		}

		if got := cfg.Get("core", "", "bare"); got != "true" {
			t.Errorf("core.bare: got %q, want %q", got, "true")
		}
		if got := cfg.Get("core", "", "filemode"); got != "false" {
			t.Errorf("core.filemode: got %q, want %q", got, "false")
		}
		if got := cfg.Get("user", "", "name"); got != "Jane Doe" {
			t.Errorf("user.name: got %q, want %q", got, "Jane Doe")
		}
		if got := cfg.Get("user", "", "email"); got != "jane@example.org" {
			t.Errorf("user.email: got %q, want %q", got, "jane@example.org")
		}
	})
}

func TestConfigSubsectionAgainstGit(t *testing.T) {
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		testRepo.Run(t, "config", "remote.origin.url", "https://example.org/repo.git")
		testRepo.Run(t, "config", "remote.origin.fetch", "+refs/heads/*:refs/remotes/origin/*")

		cfgFile := openConfig(t, testRepo)
		defer func() { _ = cfgFile.Close() }()

		cfg, err := config.ParseConfig(cfgFile)
		if err != nil {
			t.Fatalf("ParseConfig failed: %v", err)
		}

		if got := cfg.Get("remote", "origin", "url"); got != "https://example.org/repo.git" {
			t.Errorf("remote.origin.url: got %q, want %q", got, "https://example.org/repo.git")
		}
		if got := cfg.Get("remote", "origin", "fetch"); got != "+refs/heads/*:refs/remotes/origin/*" {
			t.Errorf("remote.origin.fetch: got %q, want %q", got, "+refs/heads/*:refs/remotes/origin/*")
		}
	})
}

func TestConfigMultiValueAgainstGit(t *testing.T) {
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		testRepo.Run(t, "config", "--add", "remote.origin.fetch", "+refs/heads/main:refs/remotes/origin/main")
		testRepo.Run(t, "config", "--add", "remote.origin.fetch", "+refs/heads/dev:refs/remotes/origin/dev")
		testRepo.Run(t, "config", "--add", "remote.origin.fetch", "+refs/tags/*:refs/tags/*")

		cfgFile := openConfig(t, testRepo)
		defer func() { _ = cfgFile.Close() }()

		cfg, err := config.ParseConfig(cfgFile)
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
	})
}

func TestConfigCaseInsensitiveAgainstGit(t *testing.T) {
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		testRepo.Run(t, "config", "Core.Bare", "true")
		testRepo.Run(t, "config", "CORE.FileMode", "false")

		gitVerifyBare := gitConfigGet(t, testRepo, "core.bare")
		gitVerifyFilemode := gitConfigGet(t, testRepo, "core.filemode")

		cfgFile := openConfig(t, testRepo)
		defer func() { _ = cfgFile.Close() }()

		cfg, err := config.ParseConfig(cfgFile)
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
	})
}

func TestConfigBooleanAgainstGit(t *testing.T) {
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		testRepo.Run(t, "config", "test.flag1", "true")
		testRepo.Run(t, "config", "test.flag2", "false")
		testRepo.Run(t, "config", "test.flag3", "yes")
		testRepo.Run(t, "config", "test.flag4", "no")

		cfgFile := openConfig(t, testRepo)
		defer func() { _ = cfgFile.Close() }()

		cfg, err := config.ParseConfig(cfgFile)
		if err != nil {
			t.Fatalf("ParseConfig failed: %v", err)
		}

		tests := []struct {
			key  string
			want string
		}{
			{"flag1", gitConfigGet(t, testRepo, "test.flag1")},
			{"flag2", gitConfigGet(t, testRepo, "test.flag2")},
			{"flag3", gitConfigGet(t, testRepo, "test.flag3")},
			{"flag4", gitConfigGet(t, testRepo, "test.flag4")},
		}

		for _, tt := range tests {
			if got := cfg.Get("test", "", tt.key); got != tt.want {
				t.Errorf("test.%s: got %q, want %q (from git)", tt.key, got, tt.want)
			}
		}
	})
}

func TestConfigComplexValuesAgainstGit(t *testing.T) {
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		testRepo.Run(t, "config", "test.spaced", "value with spaces")
		testRepo.Run(t, "config", "test.special", "value=with=equals")
		testRepo.Run(t, "config", "test.path", "/path/to/something")
		testRepo.Run(t, "config", "test.number", "12345")

		cfgFile := openConfig(t, testRepo)
		defer func() { _ = cfgFile.Close() }()

		cfg, err := config.ParseConfig(cfgFile)
		if err != nil {
			t.Fatalf("ParseConfig failed: %v", err)
		}

		tests := []string{"spaced", "special", "path", "number"}
		for _, key := range tests {
			want := gitConfigGet(t, testRepo, "test."+key)
			if got := cfg.Get("test", "", key); got != want {
				t.Errorf("test.%s: got %q, want %q (from git)", key, got, want)
			}
		}
	})
}

func TestConfigEntriesAgainstGit(t *testing.T) {
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		testRepo.Run(t, "config", "core.bare", "true")
		testRepo.Run(t, "config", "core.filemode", "false")
		testRepo.Run(t, "config", "user.name", "Test User")

		cfgFile := openConfig(t, testRepo)
		defer func() { _ = cfgFile.Close() }()

		cfg, err := config.ParseConfig(cfgFile)
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

			gitValue := gitConfigGet(t, testRepo, key)
			if entry.Value != gitValue {
				t.Errorf("entry %s: got value %q, git has %q", key, entry.Value, gitValue)
			}
		}
	})
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
			_, err := config.ParseConfig(r)
			if err == nil {
				t.Errorf("expected error for %s", tt.name)
			}
		})
	}
}
