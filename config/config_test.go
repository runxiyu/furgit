package config_test

import (
	"bytes"
	"os"
	"os/exec"
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

func gitConfigGetE(testRepo *testgit.TestRepo, key string) (string, error) {
	//nolint:noctx
	cmd := exec.Command("git", "config", "--get", key) //#nosec G204
	cmd.Dir = testRepo.Dir()
	cmd.Env = testRepo.Env()
	out, err := cmd.CombinedOutput()

	return strings.TrimSpace(string(out)), err
}

func lookupValue(cfg *config.Config, section, subsection, key string) string {
	result := cfg.Lookup(section, subsection, key)
	if result.Kind == config.ValueMissing {
		return ""
	}

	return result.Value
}

func lookupAllValues(cfg *config.Config, section, subsection, key string) []string {
	results := cfg.LookupAll(section, subsection, key)

	values := make([]string, 0, len(results))
	for _, result := range results {
		values = append(values, result.Value)
	}

	return values
}

func TestConfigAgainstGit(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
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

		if got := lookupValue(cfg, "core", "", "bare"); got != "true" {
			t.Errorf("core.bare: got %q, want %q", got, "true")
		}

		if got := lookupValue(cfg, "core", "", "filemode"); got != "false" {
			t.Errorf("core.filemode: got %q, want %q", got, "false")
		}

		if got := lookupValue(cfg, "user", "", "name"); got != "Jane Doe" {
			t.Errorf("user.name: got %q, want %q", got, "Jane Doe")
		}

		if got := lookupValue(cfg, "user", "", "email"); got != "jane@example.org" {
			t.Errorf("user.email: got %q, want %q", got, "jane@example.org")
		}
	})
}

func TestConfigSubsectionAgainstGit(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		testRepo.Run(t, "config", "remote.origin.url", "https://example.org/repo.git")
		testRepo.Run(t, "config", "remote.origin.fetch", "+refs/heads/*:refs/remotes/origin/*")

		cfgFile := openConfig(t, testRepo)

		defer func() { _ = cfgFile.Close() }()

		cfg, err := config.ParseConfig(cfgFile)
		if err != nil {
			t.Fatalf("ParseConfig failed: %v", err)
		}

		if got := lookupValue(cfg, "remote", "origin", "url"); got != "https://example.org/repo.git" {
			t.Errorf("remote.origin.url: got %q, want %q", got, "https://example.org/repo.git")
		}

		if got := lookupValue(cfg, "remote", "origin", "fetch"); got != "+refs/heads/*:refs/remotes/origin/*" {
			t.Errorf("remote.origin.fetch: got %q, want %q", got, "+refs/heads/*:refs/remotes/origin/*")
		}
	})
}

func TestConfigMultiValueAgainstGit(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
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

		fetches := lookupAllValues(cfg, "remote", "origin", "fetch")
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
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
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

		if got := lookupValue(cfg, "core", "", "bare"); got != gitVerifyBare {
			t.Errorf("core.bare: got %q, want %q (from git)", got, gitVerifyBare)
		}

		if got := lookupValue(cfg, "CORE", "", "BARE"); got != gitVerifyBare {
			t.Errorf("CORE.BARE: got %q, want %q (from git)", got, gitVerifyBare)
		}

		if got := lookupValue(cfg, "core", "", "filemode"); got != gitVerifyFilemode {
			t.Errorf("core.filemode: got %q, want %q (from git)", got, gitVerifyFilemode)
		}
	})
}

func TestConfigBooleanAgainstGit(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
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
			if got := lookupValue(cfg, "test", "", tt.key); got != tt.want {
				t.Errorf("test.%s: got %q, want %q (from git)", tt.key, got, tt.want)
			}
		}
	})
}

func TestConfigLookupKindsAndBool(t *testing.T) {
	t.Parallel()

	cfgText := "[test]\nnovalue\nempty =\ntruthy = yes\nnumeric = -2\nleadspace = \" 1\"\nleadtab = \"\t-2\"\nksuffix = 1k\nhex = 0x10\nmaxi32 = 2147483647\ntoobig = 2147483648\ntoosmall = -2147483649\nbadnum = \" 2x\"\n"

	cfg, err := config.ParseConfig(strings.NewReader(cfgText))
	if err != nil {
		t.Fatalf("ParseConfig failed: %v", err)
	}

	novalue := cfg.Lookup("test", "", "novalue")
	if novalue.Kind != config.ValueValueless {
		t.Fatalf("novalue kind: got %v, want %v", novalue.Kind, config.ValueValueless)
	}

	novalueBool, err := novalue.Bool()
	if err != nil || !novalueBool {
		t.Fatalf("novalue bool: got (%v, %v), want (true, nil)", novalueBool, err)
	}

	empty := cfg.Lookup("test", "", "empty")
	if empty.Kind != config.ValueString || empty.Value != "" {
		t.Fatalf("empty: got (%v, %q), want (%v, %q)", empty.Kind, empty.Value, config.ValueString, "")
	}

	emptyBool, err := empty.Bool()
	if err != nil || emptyBool {
		t.Fatalf("empty bool: got (%v, %v), want (false, nil)", emptyBool, err)
	}

	truthyBool, err := cfg.Lookup("test", "", "truthy").Bool()
	if err != nil || !truthyBool {
		t.Fatalf("truthy bool: got (%v, %v), want (true, nil)", truthyBool, err)
	}

	numericBool, err := cfg.Lookup("test", "", "numeric").Bool()
	if err != nil || !numericBool {
		t.Fatalf("numeric bool: got (%v, %v), want (true, nil)", numericBool, err)
	}

	leadspaceBool, err := cfg.Lookup("test", "", "leadspace").Bool()
	if err != nil || !leadspaceBool {
		t.Fatalf("leadspace bool: got (%v, %v), want (true, nil)", leadspaceBool, err)
	}

	leadtabBool, err := cfg.Lookup("test", "", "leadtab").Bool()
	if err != nil || !leadtabBool {
		t.Fatalf("leadtab bool: got (%v, %v), want (true, nil)", leadtabBool, err)
	}

	ksuffixBool, err := cfg.Lookup("test", "", "ksuffix").Bool()
	if err != nil || !ksuffixBool {
		t.Fatalf("ksuffix bool: got (%v, %v), want (true, nil)", ksuffixBool, err)
	}

	maxi32Bool, err := cfg.Lookup("test", "", "maxi32").Bool()
	if err != nil || !maxi32Bool {
		t.Fatalf("maxi32 bool: got (%v, %v), want (true, nil)", maxi32Bool, err)
	}

	_, err = cfg.Lookup("test", "", "toobig").Bool()
	if err == nil {
		t.Fatal("toobig bool: expected error")
	}

	_, err = cfg.Lookup("test", "", "toosmall").Bool()
	if err == nil {
		t.Fatal("toosmall bool: expected error")
	}

	_, err = cfg.Lookup("test", "", "badnum").Bool()
	if err == nil {
		t.Fatal("badnum bool: expected error")
	}

	_, err = novalue.String()
	if err == nil {
		t.Fatal("novalue string: expected error")
	}

	emptyString, err := empty.String()
	if err != nil || emptyString != "" {
		t.Fatalf("empty string: got (%q, %v), want (%q, nil)", emptyString, err, "")
	}

	numericInt, err := cfg.Lookup("test", "", "numeric").Int()
	if err != nil || numericInt != -2 {
		t.Fatalf("numeric int: got (%v, %v), want (-2, nil)", numericInt, err)
	}

	ksuffixInt, err := cfg.Lookup("test", "", "ksuffix").Int()
	if err != nil || ksuffixInt != 1024 {
		t.Fatalf("ksuffix int: got (%v, %v), want (1024, nil)", ksuffixInt, err)
	}

	hexInt64, err := cfg.Lookup("test", "", "hex").Int64()
	if err != nil || hexInt64 != 16 {
		t.Fatalf("hex int64: got (%v, %v), want (16, nil)", hexInt64, err)
	}

	_, err = cfg.Lookup("test", "", "badnum").Int()
	if err == nil {
		t.Fatal("badnum int: expected error")
	}

	missing := cfg.Lookup("test", "", "missing")
	if missing.Kind != config.ValueMissing {
		t.Fatalf("missing kind: got %v, want %v", missing.Kind, config.ValueMissing)
	}

	_, err = missing.Bool()
	if err == nil {
		t.Fatal("missing bool: expected error")
	}

	_, err = missing.Int()
	if err == nil {
		t.Fatal("missing int: expected error")
	}

	_, err = missing.String()
	if err == nil {
		t.Fatal("missing string: expected error")
	}
}

func TestConfigComplexValuesAgainstGit(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
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
			if got := lookupValue(cfg, "test", "", key); got != want {
				t.Errorf("test.%s: got %q, want %q (from git)", key, got, want)
			}
		}
	})
}

func TestConfigEntriesAgainstGit(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
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
	t.Parallel()

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
			t.Parallel()

			r := strings.NewReader(tt.config)

			_, err := config.ParseConfig(r)
			if err == nil {
				t.Errorf("expected error for %s", tt.name)
			}
		})
	}
}

func TestConfigEOFAfterKeyAgainstGit(t *testing.T) { //nolint:dupl
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		cfgPath := filepath.Join(testRepo.Dir(), "config")

		cfgData := []byte("[Core]BAre")

		err := os.WriteFile(cfgPath, cfgData, 0o600)
		if err != nil {
			t.Fatalf("failed to write config: %v", err)
		}

		gitValue, gitErr := gitConfigGetE(testRepo, "Core.BAre")
		furConfig, furErr := config.ParseConfig(bytes.NewReader(cfgData))

		if (gitErr == nil) != (furErr == nil) {
			t.Fatalf("git: %v\nfur: %v", gitErr, furErr)
		}

		if furErr != nil {
			return
		}

		if got := lookupValue(furConfig, "Core", "", "BAre"); got != gitValue {
			t.Fatalf("git: %q\nfur: %q", gitValue, got)
		}
	})
}

func TestConfigNULValueAgainstGit(t *testing.T) { //nolint:dupl
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		cfgPath := filepath.Join(testRepo.Dir(), "config")

		cfgData := []byte("[Core]BAre=\x00")

		err := os.WriteFile(cfgPath, cfgData, 0o600)
		if err != nil {
			t.Fatalf("failed to write config: %v", err)
		}

		gitValue, gitErr := gitConfigGetE(testRepo, "Core.BAre")
		furConfig, furErr := config.ParseConfig(bytes.NewReader(cfgData))

		if (gitErr == nil) != (furErr == nil) {
			t.Fatalf("git: %v\nfur: %v", gitErr, furErr)
		}

		if furErr != nil {
			return
		}

		if got := lookupValue(furConfig, "Core", "", "BAre"); got != gitValue {
			t.Fatalf("git: %q\nfur: %q", gitValue, got)
		}
	})
}

func TestConfigCarriageReturnSeparatorAgainstGit(t *testing.T) { //nolint:dupl
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		cfgPath := filepath.Join(testRepo.Dir(), "config")

		cfgData := []byte("[Core \"sub\"]\rBAre")

		err := os.WriteFile(cfgPath, cfgData, 0o600)
		if err != nil {
			t.Fatalf("failed to write config: %v", err)
		}

		gitValue, gitErr := gitConfigGetE(testRepo, "Core.sub.BAre")
		furConfig, furErr := config.ParseConfig(bytes.NewReader(cfgData))

		if (gitErr == nil) != (furErr == nil) {
			t.Fatalf("git: %v\nfur: %v", gitErr, furErr)
		}

		if furErr != nil {
			return
		}

		if got := lookupValue(furConfig, "Core", "sub", "BAre"); got != gitValue {
			t.Fatalf("git: %q\nfur: %q", gitValue, got)
		}
	})
}

func FuzzConfig(f *testing.F) {
	f.Add([]byte("[core]\nbare = true"), "core.bare")
	f.Add([]byte("[core]\nbare = true\n[core/invalid]"), "core.bare")
	f.Add([]byte("[core \"sub\"]\nbare = true"), "core.sub.bare")

	type fuzzRepoState struct {
		repo    *testgit.TestRepo
		cfgPath string
	}

	repos := make(map[objectid.Algorithm]fuzzRepoState, len(objectid.SupportedAlgorithms()))
	for _, algo := range objectid.SupportedAlgorithms() {
		testRepo := testgit.NewRepo(f, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		repos[algo] = fuzzRepoState{
			repo:    testRepo,
			cfgPath: filepath.Join(testRepo.Dir(), "config"),
		}
	}

	f.Fuzz(func(t *testing.T, cfgData []byte, gitKey string) {
		for _, algo := range objectid.SupportedAlgorithms() {
			state, ok := repos[algo]
			if !ok {
				t.Fatalf("missing fuzz repo state for %v", algo)
			}

			err := os.WriteFile(state.cfgPath, cfgData, 0o600)
			if err != nil {
				t.Fatalf("failed to write config: %v", err)
			}

			gitValue, gitErr := gitConfigGetE(state.repo, gitKey)

			furConfig, furErr := config.ParseConfig(bytes.NewReader(cfgData))
			if furErr == nil && furConfig == nil {
				t.Fatalf("ParseConfig returned nil config with nil error")
			}

			sameErr := (gitErr == nil) == (furErr == nil)
			if !sameErr {
				if furErr == nil {
					return
				}

				t.Fatalf("git: %v\nfur: %v", gitErr, furErr)
			}

			if furErr == nil {
				parts := strings.SplitN(gitKey, ".", 3)
				furSection := parts[0]

				var furSubsection, furKey string

				switch len(parts) {
				case 1:
				case 2:
					furKey = parts[1]
				case 3:
					furSubsection = parts[1]
					furKey = parts[2]
				default:
					t.Fatalf("unexpected split(%q): %v", gitKey, parts)
				}

				furValue := lookupValue(furConfig, furSection, furSubsection, furKey)
				if gitValue != furValue {
					t.Fatalf(
						"key: %v (%v.%v.%v)\ngit: %q\nfur: %q",
						gitKey, furSection, furSubsection, furKey, gitValue, furValue,
					)
				}
			}
		}
	})
}
