package config_test

import (
	"bytes"
	"strings"
	"testing"

	"lindenii.org/go/furgit/config"
	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/object/id"
)

func TestConfig(t *testing.T) {
	t.Parallel()

	testRepo, err := testgit.NewRepo(t, testgit.RepoOptions{})
	if err != nil {
		t.Fatalf("NewRepo: %v", err)
	}

	err = testRepo.ConfigSet(t, "test.enabled", "true")
	if err != nil {
		t.Fatalf("set test.enabled: %v", err)
	}

	err = testRepo.ConfigSet(t, "test.mode", "false")
	if err != nil {
		t.Fatalf("set test.mode: %v", err)
	}

	err = testRepo.ConfigSet(t, "test.name", "Jane Doe")
	if err != nil {
		t.Fatalf("set test.name: %v", err)
	}

	err = testRepo.ConfigSet(t, "test.email", "jane@example.org")
	if err != nil {
		t.Fatalf("set test.email: %v", err)
	}

	cfgFile, err := testRepo.Root(t).Open(".git/config")
	if err != nil {
		t.Fatalf("open config: %v", err)
	}

	defer func() { _ = cfgFile.Close() }()

	cfg, err := config.Parse(cfgFile)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	got, err := cfg.Lookup("test", "", "enabled").String()
	if err != nil {
		t.Fatalf("test.enabled: %v", err)
	}

	if got != "true" {
		t.Errorf("test.enabled: got %q, want %q", got, "true")
	}

	got, err = cfg.Lookup("test", "", "mode").String()
	if err != nil {
		t.Fatalf("test.mode: %v", err)
	}

	if got != "false" {
		t.Errorf("test.mode: got %q, want %q", got, "false")
	}

	got, err = cfg.Lookup("test", "", "name").String()
	if err != nil {
		t.Fatalf("test.name: %v", err)
	}

	if got != "Jane Doe" {
		t.Errorf("test.name: got %q, want %q", got, "Jane Doe")
	}

	got, err = cfg.Lookup("test", "", "email").String()
	if err != nil {
		t.Fatalf("test.email: %v", err)
	}

	if got != "jane@example.org" {
		t.Errorf("test.email: got %q, want %q", got, "jane@example.org")
	}
}

func TestConfigSubsection(t *testing.T) {
	t.Parallel()

	testRepo, err := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: id.ObjectFormatSHA256})
	if err != nil {
		t.Fatalf("NewRepo: %v", err)
	}

	err = testRepo.ConfigSet(t, "test.origin.url", "https://example.org/repo.git")
	if err != nil {
		t.Fatalf("set test.origin.url: %v", err)
	}

	err = testRepo.ConfigSet(t, "test.origin.fetch", "+refs/heads/*:refs/remotes/origin/*")
	if err != nil {
		t.Fatalf("set test.origin.fetch: %v", err)
	}

	cfgFile, err := testRepo.Root(t).Open(".git/config")
	if err != nil {
		t.Fatalf("open config: %v", err)
	}

	defer func() { _ = cfgFile.Close() }()

	cfg, err := config.Parse(cfgFile)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	got, err := cfg.Lookup("test", "origin", "url").String()
	if err != nil {
		t.Fatalf("test.origin.url: %v", err)
	}

	if got != "https://example.org/repo.git" {
		t.Errorf("test.origin.url: got %q, want %q", got, "https://example.org/repo.git")
	}

	got, err = cfg.Lookup("test", "origin", "fetch").String()
	if err != nil {
		t.Fatalf("test.origin.fetch: %v", err)
	}

	if got != "+refs/heads/*:refs/remotes/origin/*" {
		t.Errorf("test.origin.fetch: got %q, want %q", got, "+refs/heads/*:refs/remotes/origin/*")
	}
}

func TestConfigMultiValue(t *testing.T) {
	t.Parallel()

	testRepo, err := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: id.ObjectFormatSHA256})
	if err != nil {
		t.Fatalf("NewRepo: %v", err)
	}

	err = testRepo.ConfigAdd(t, "test.origin.fetch", "+refs/heads/main:refs/remotes/origin/main")
	if err != nil {
		t.Fatalf("add test.origin.fetch: %v", err)
	}

	err = testRepo.ConfigAdd(t, "test.origin.fetch", "+refs/heads/dev:refs/remotes/origin/dev")
	if err != nil {
		t.Fatalf("add test.origin.fetch: %v", err)
	}

	err = testRepo.ConfigAdd(t, "test.origin.fetch", "+refs/tags/*:refs/tags/*")
	if err != nil {
		t.Fatalf("add test.origin.fetch: %v", err)
	}

	cfgFile, err := testRepo.Root(t).Open(".git/config")
	if err != nil {
		t.Fatalf("open config: %v", err)
	}

	defer func() { _ = cfgFile.Close() }()

	cfg, err := config.Parse(cfgFile)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	fetches := cfg.LookupAll("test", "origin", "fetch")
	if len(fetches) != 3 {
		t.Fatalf("expected 3 fetch values, got %d", len(fetches))
	}

	expected := []string{
		"+refs/heads/main:refs/remotes/origin/main",
		"+refs/heads/dev:refs/remotes/origin/dev",
		"+refs/tags/*:refs/tags/*",
	}
	for i, want := range expected {
		got, err := fetches[i].String()
		if err != nil {
			t.Fatalf("fetch[%d]: %v", i, err)
		}

		if got != want {
			t.Errorf("fetch[%d]: got %q, want %q", i, got, want)
		}
	}
}

func TestConfigLookupLastWins(t *testing.T) {
	t.Parallel()

	testRepo, err := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: id.ObjectFormatSHA256})
	if err != nil {
		t.Fatalf("NewRepo: %v", err)
	}

	for _, value := range []string{"first", "second", "third"} {
		err = testRepo.ConfigAdd(t, "test.single", value)
		if err != nil {
			t.Fatalf("add test.single: %v", err)
		}
	}

	cfgFile, err := testRepo.Root(t).Open(".git/config")
	if err != nil {
		t.Fatalf("open config: %v", err)
	}

	defer func() { _ = cfgFile.Close() }()

	cfg, err := config.Parse(cfgFile)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	got, err := cfg.Lookup("test", "", "single").String()
	if err != nil {
		t.Fatalf("test.single: %v", err)
	}

	want, err := testRepo.ConfigGet(t, "test.single")
	if err != nil {
		t.Fatalf("ConfigGet: %v", err)
	}

	want = strings.TrimSuffix(want, "\n")
	if got != want {
		t.Errorf("test.single: got %q, want %q", got, want)
	}

	if got != "third" {
		t.Errorf("test.single: got %q, want %q", got, "third")
	}
}

func TestConfigMerge(t *testing.T) {
	t.Parallel()

	first, err := config.Parse(strings.NewReader("[test]\n\tkey = one\n\tonly = first\n"))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	second, err := config.Parse(strings.NewReader("[test]\n\tkey = two\n"))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	merged := config.Merge(first, second)

	got, err := merged.Lookup("test", "", "key").String()
	if err != nil {
		t.Fatalf("test.key: %v", err)
	}

	if got != "two" {
		t.Errorf("test.key: got %q, want %q", got, "two")
	}

	got, err = merged.Lookup("test", "", "only").String()
	if err != nil {
		t.Fatalf("test.only: %v", err)
	}

	if got != "first" {
		t.Errorf("test.only: got %q, want %q", got, "first")
	}

	all := merged.LookupAll("test", "", "key")
	if len(all) != 2 {
		t.Fatalf("expected 2 values, got %d", len(all))
	}

	if all[0].Value != "one" || all[1].Value != "two" {
		t.Errorf("LookupAll = %q, %q, want %q, %q", all[0].Value, all[1].Value, "one", "two")
	}

	// Merging must not disturb the sources.
	got, err = first.Lookup("test", "", "key").String()
	if err != nil {
		t.Fatalf("first test.key: %v", err)
	}

	if got != "one" {
		t.Errorf("first test.key: got %q, want %q", got, "one")
	}
}

func TestConfigLookupLastWinsOverValueless(t *testing.T) {
	t.Parallel()

	cfg, err := config.Parse(strings.NewReader("[test]\n\tflag\n\tflag = false\n"))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	got, err := cfg.Lookup("test", "", "flag").String()
	if err != nil {
		t.Fatalf("test.flag: %v", err)
	}

	if got != "false" {
		t.Errorf("test.flag: got %q, want %q", got, "false")
	}
}

func TestConfigCaseInsensitive(t *testing.T) {
	t.Parallel()

	testRepo, err := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: id.ObjectFormatSHA256})
	if err != nil {
		t.Fatalf("NewRepo: %v", err)
	}

	err = testRepo.ConfigSet(t, "Test.Flag", "true")
	if err != nil {
		t.Fatalf("set Test.Flag: %v", err)
	}

	err = testRepo.ConfigSet(t, "TEST.Mode", "false")
	if err != nil {
		t.Fatalf("set TEST.Mode: %v", err)
	}

	gitVerifyFlag, err := testRepo.ConfigGet(t, "test.flag")
	if err != nil {
		t.Fatalf("get test.flag: %v", err)
	}

	gitVerifyMode, err := testRepo.ConfigGet(t, "test.mode")
	if err != nil {
		t.Fatalf("get test.mode: %v", err)
	}

	gitVerifyFlag = strings.TrimSuffix(gitVerifyFlag, "\n")
	gitVerifyMode = strings.TrimSuffix(gitVerifyMode, "\n")

	cfgFile, err := testRepo.Root(t).Open(".git/config")
	if err != nil {
		t.Fatalf("open config: %v", err)
	}

	defer func() { _ = cfgFile.Close() }()

	cfg, err := config.Parse(cfgFile)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	got, err := cfg.Lookup("test", "", "flag").String()
	if err != nil {
		t.Fatalf("test.flag: %v", err)
	}

	if got != gitVerifyFlag {
		t.Errorf("test.flag: got %q, want %q (from git)", got, gitVerifyFlag)
	}

	got, err = cfg.Lookup("TEST", "", "FLAG").String()
	if err != nil {
		t.Fatalf("TEST.FLAG: %v", err)
	}

	if got != gitVerifyFlag {
		t.Errorf("TEST.FLAG: got %q, want %q (from git)", got, gitVerifyFlag)
	}

	got, err = cfg.Lookup("test", "", "mode").String()
	if err != nil {
		t.Fatalf("test.mode: %v", err)
	}

	if got != gitVerifyMode {
		t.Errorf("test.mode: got %q, want %q (from git)", got, gitVerifyMode)
	}
}

func TestConfigBoolean(t *testing.T) {
	t.Parallel()

	testRepo, err := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: id.ObjectFormatSHA256})
	if err != nil {
		t.Fatalf("NewRepo: %v", err)
	}

	err = testRepo.ConfigSet(t, "test.flag1", "true")
	if err != nil {
		t.Fatalf("set test.flag1: %v", err)
	}

	err = testRepo.ConfigSet(t, "test.flag2", "false")
	if err != nil {
		t.Fatalf("set test.flag2: %v", err)
	}

	err = testRepo.ConfigSet(t, "test.flag3", "yes")
	if err != nil {
		t.Fatalf("set test.flag3: %v", err)
	}

	err = testRepo.ConfigSet(t, "test.flag4", "no")
	if err != nil {
		t.Fatalf("set test.flag4: %v", err)
	}

	cfgFile, err := testRepo.Root(t).Open(".git/config")
	if err != nil {
		t.Fatalf("open config: %v", err)
	}

	defer func() { _ = cfgFile.Close() }()

	cfg, err := config.Parse(cfgFile)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	tests := make([]struct {
		key  string
		want string
	}, 0, 4)
	for _, key := range []string{"flag1", "flag2", "flag3", "flag4"} {
		want, err := testRepo.ConfigGet(t, "test."+key)
		if err != nil {
			t.Fatalf("get test.%s: %v", key, err)
		}

		tests = append(tests, struct {
			key  string
			want string
		}{
			key:  key,
			want: strings.TrimSuffix(want, "\n"),
		})
	}

	for _, tt := range tests {
		got, err := cfg.Lookup("test", "", tt.key).String()
		if err != nil {
			t.Fatalf("test.%s: %v", tt.key, err)
		}

		if got != tt.want {
			t.Errorf("test.%s: got %q, want %q (from git)", tt.key, got, tt.want)
		}
	}
}

func TestConfigLookupKindsAndBool(t *testing.T) {
	t.Parallel()

	cfgText := `[test]
novalue
empty =
truthy = yes
numeric = -2
leadspace = " 1"
leadtab = "	-2"
ksuffix = 1k
hex = 0x10
maxi32 = 2147483647
toobig = 2147483648
toosmall = -2147483649
badnum = " 2x"
`

	cfg, err := config.Parse(strings.NewReader(cfgText))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	novalue := cfg.Lookup("test", "", "novalue")
	if novalue.Kind != config.KindValueless {
		t.Fatalf("novalue kind: got %v, want %v", novalue.Kind, config.KindValueless)
	}

	novalueBool, err := novalue.Bool()
	if err != nil || !novalueBool {
		t.Fatalf("novalue bool: got (%v, %v), want (true, nil)", novalueBool, err)
	}

	empty := cfg.Lookup("test", "", "empty")
	if empty.Kind != config.KindString || empty.Value != "" {
		t.Fatalf("empty: got (%v, %q), want (%v, %q)", empty.Kind, empty.Value, config.KindString, "")
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
	if missing.Kind != config.KindMissing {
		t.Fatalf("missing kind: got %v, want %v", missing.Kind, config.KindMissing)
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

func TestConfigComplexValues(t *testing.T) {
	t.Parallel()

	testRepo, err := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: id.ObjectFormatSHA256})
	if err != nil {
		t.Fatalf("NewRepo: %v", err)
	}

	err = testRepo.ConfigSet(t, "test.spaced", "value with spaces")
	if err != nil {
		t.Fatalf("set test.spaced: %v", err)
	}

	err = testRepo.ConfigSet(t, "test.special", "value=with=equals")
	if err != nil {
		t.Fatalf("set test.special: %v", err)
	}

	err = testRepo.ConfigSet(t, "test.path", "/path/to/something")
	if err != nil {
		t.Fatalf("set test.path: %v", err)
	}

	err = testRepo.ConfigSet(t, "test.number", "12345")
	if err != nil {
		t.Fatalf("set test.number: %v", err)
	}

	cfgFile, err := testRepo.Root(t).Open(".git/config")
	if err != nil {
		t.Fatalf("open config: %v", err)
	}

	defer func() { _ = cfgFile.Close() }()

	cfg, err := config.Parse(cfgFile)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	tests := []string{"spaced", "special", "path", "number"}
	for _, key := range tests {
		want, err := testRepo.ConfigGet(t, "test."+key)
		if err != nil {
			t.Fatalf("get test.%s: %v", key, err)
		}

		want = strings.TrimSuffix(want, "\n")

		got, err := cfg.Lookup("test", "", key).String()
		if err != nil {
			t.Fatalf("test.%s: %v", key, err)
		}

		if got != want {
			t.Errorf("test.%s: got %q, want %q (from git)", key, got, want)
		}
	}
}

func TestConfigEntries(t *testing.T) {
	t.Parallel()

	testRepo, err := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: id.ObjectFormatSHA256})
	if err != nil {
		t.Fatalf("NewRepo: %v", err)
	}

	err = testRepo.ConfigSet(t, "test.enabled", "true")
	if err != nil {
		t.Fatalf("set test.enabled: %v", err)
	}

	err = testRepo.ConfigSet(t, "test.mode", "false")
	if err != nil {
		t.Fatalf("set test.mode: %v", err)
	}

	err = testRepo.ConfigSet(t, "test.name", "Test User")
	if err != nil {
		t.Fatalf("set test.name: %v", err)
	}

	cfgFile, err := testRepo.Root(t).Open(".git/config")
	if err != nil {
		t.Fatalf("open config: %v", err)
	}

	defer func() { _ = cfgFile.Close() }()

	cfg, err := config.Parse(cfgFile)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	entries := cfg.Entries()
	if len(entries) < 3 {
		t.Errorf("expected at least 3 entries, got %d", len(entries))
	}

	for _, entry := range entries {
		key := entry.Section + "." + entry.Key
		if entry.Subsection != "" {
			key = entry.Section + "." + entry.Subsection + "." + entry.Key
		}

		gitValue, err := testRepo.ConfigGet(t, key)
		if err != nil {
			t.Fatalf("get %s: %v", key, err)
		}

		gitValue = strings.TrimSuffix(gitValue, "\n")
		if entry.Value != gitValue {
			t.Errorf("entry %s: got value %q, git has %q", key, entry.Value, gitValue)
		}
	}
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

			_, err := config.Parse(r)
			if err == nil {
				t.Errorf("expected error for %s", tt.name)
			}
		})
	}
}

func TestConfigRawBytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		cfgData    []byte
		gitKey     string
		section    string
		subsection string
		key        string
	}{
		{
			name:       "EOF after empty value",
			cfgData:    []byte("[Test]Flag="),
			gitKey:     "Test.Flag",
			section:    "Test",
			subsection: "",
			key:        "Flag",
		},
		{
			name:       "NUL value",
			cfgData:    []byte("[Test]Flag=\x00"),
			gitKey:     "Test.Flag",
			section:    "Test",
			subsection: "",
			key:        "Flag",
		},
		{
			name:       "carriage return separator",
			cfgData:    []byte("[Test \"sub\"]\rFlag="),
			gitKey:     "Test.sub.Flag",
			section:    "Test",
			subsection: "sub",
			key:        "Flag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			testRepo, err := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: id.ObjectFormatSHA256})
			if err != nil {
				t.Fatalf("NewRepo: %v", err)
			}

			err = testRepo.Root(t).WriteFile(".git/config", tt.cfgData, 0o600)
			if err != nil {
				t.Fatalf("write config: %v", err)
			}

			gitValue, gitErr := testRepo.ConfigGet(t, tt.gitKey)
			furConfig, furErr := config.Parse(bytes.NewReader(tt.cfgData))

			if (gitErr == nil) != (furErr == nil) {
				t.Fatalf("git: %v\nfur: %v", gitErr, furErr)
			}

			if furErr != nil {
				return
			}

			gitValue = strings.TrimSuffix(gitValue, "\n")

			got, err := furConfig.Lookup(tt.section, tt.subsection, tt.key).String()
			if err != nil {
				t.Fatalf("%s: %v", tt.gitKey, err)
			}

			if got != gitValue {
				t.Fatalf("git: %q\nfur: %q", gitValue, got)
			}
		})
	}
}

func FuzzConfig(f *testing.F) {
	f.Add([]byte("[test]\nflag = true"), "test.flag")
	f.Add([]byte("[test]\nflag = true\n[core/invalid]"), "test.flag")
	f.Add([]byte("[test \"sub\"]\nflag = true"), "test.sub.flag")

	testRepo, err := testgit.NewRepo(f, testgit.RepoOptions{ObjectFormat: id.ObjectFormatSHA256})
	if err != nil {
		f.Fatalf("NewRepo: %v", err)
	}

	f.Fuzz(func(t *testing.T, cfgData []byte, gitKey string) {
		err := testRepo.Root(t).WriteFile(".git/config", cfgData, 0o600)
		if err != nil {
			t.Fatalf("write config: %v", err)
		}

		gitValue, gitErr := testRepo.ConfigGet(t, gitKey)

		furConfig, furErr := config.Parse(bytes.NewReader(cfgData))
		if furErr == nil && furConfig == nil {
			t.Fatalf("Parse returned nil config with nil error")
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

			gitValue = strings.TrimSuffix(gitValue, "\n")

			furValue, err := furConfig.Lookup(furSection, furSubsection, furKey).String()
			if err != nil {
				t.Fatalf("%s: %v", gitKey, err)
			}

			if gitValue != furValue {
				t.Fatalf(
					"key: %v (%v.%v.%v)\ngit: %q\nfur: %q",
					gitKey, furSection, furSubsection, furKey, gitValue, furValue,
				)
			}
		}
	})
}
