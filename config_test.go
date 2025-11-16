package furgit

import (
	"strings"
	"testing"
)

func TestParseConfigSimple(t *testing.T) {
	input := `
[core]
	repositoryformatversion = 0
	filemode = true
	bare = false

[user]
	name = Alice Example
	email = alice@example.com
`
	cfg, err := ParseConfig(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseConfig error: %v", err)
	}

	if got := cfg.Get("core", "", "repositoryformatversion"); got != "0" {
		t.Errorf("core.repositoryformatversion = %q, want %q", got, "0")
	}
	if got := cfg.Get("core", "", "filemode"); got != "true" {
		t.Errorf("core.filemode = %q, want %q", got, "true")
	}
	if got := cfg.Get("user", "", "name"); got != "Alice Example" {
		t.Errorf("user.name = %q, want %q", got, "Alice Example")
	}
	if got := cfg.Get("user", "", "email"); got != "alice@example.com" {
		t.Errorf("user.email = %q, want %q", got, "alice@example.com")
	}
}

func TestParseConfigSubsection(t *testing.T) {
	input := `
[remote "origin"]
	url = https://villosa.example.org/group1/group2//repos/repo
	fetch = +refs/heads/*:refs/remotes/origin/*
`
	cfg, err := ParseConfig(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseConfig error: %v", err)
	}

	if got := cfg.Get("remote", "origin", "url"); got != "https://villosa.example.org/group1/group2//repos/repo" {
		t.Errorf("remote.origin.url = %q, want %q", got, "https://villosa.example.org/group1/group2//repos/repo")
	}
	if got := cfg.Get("remote", "origin", "fetch"); got != "+refs/heads/*:refs/remotes/origin/*" {
		t.Errorf("remote.origin.fetch = %q, want %q", got, "+refs/heads/*:refs/remotes/origin/*")
	}
}

func TestParseConfigCaseInsensitive(t *testing.T) {
	input := `
[Core]
	FileMode = true
`
	cfg, err := ParseConfig(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseConfig error: %v", err)
	}

	if got := cfg.Get("core", "", "filemode"); got != "true" {
		t.Errorf("core.filemode = %q, want %q", got, "true")
	}
	if got := cfg.Get("CORE", "", "FILEMODE"); got != "true" {
		t.Errorf("CORE.FILEMODE = %q, want %q", got, "true")
	}
}

func TestParseConfigBooleanKeys(t *testing.T) {
	input := `
[core]
	bare
	ignorecase
`
	cfg, err := ParseConfig(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseConfig error: %v", err)
	}

	if got := cfg.Get("core", "", "bare"); got != "true" {
		t.Errorf("core.bare = %q, want %q", got, "true")
	}
	if got := cfg.Get("core", "", "ignorecase"); got != "true" {
		t.Errorf("core.ignorecase = %q, want %q", got, "true")
	}
}

func TestParseConfigQuotedValues(t *testing.T) {
	input := `
[user]
	name = "Bob Smith"
	comment = "Has a \"quoted\" word"
`
	cfg, err := ParseConfig(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseConfig error: %v", err)
	}

	if got := cfg.Get("user", "", "name"); got != "Bob Smith" {
		t.Errorf("user.name = %q, want %q", got, "Bob Smith")
	}
	if got := cfg.Get("user", "", "comment"); got != `Has a "quoted" word` {
		t.Errorf("user.comment = %q, want %q", got, `Has a "quoted" word`)
	}
}

func TestParseConfigEscapeSequences(t *testing.T) {
	input := `
[test]
	newline = "line1\nline2"
	tab = "col1\tcol2"
	backslash = "path\\to\\file"
`
	cfg, err := ParseConfig(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseConfig error: %v", err)
	}

	if got := cfg.Get("test", "", "newline"); got != "line1\nline2" {
		t.Errorf("test.newline = %q, want %q", got, "line1\nline2")
	}
	if got := cfg.Get("test", "", "tab"); got != "col1\tcol2" {
		t.Errorf("test.tab = %q, want %q", got, "col1\tcol2")
	}
	if got := cfg.Get("test", "", "backslash"); got != "path\\to\\file" {
		t.Errorf("test.backslash = %q, want %q", got, "path\\to\\file")
	}
}

func TestParseConfigComments(t *testing.T) {
	input := `
# This is a comment
; This is also a comment
[core]
	# Comment in section
	bare = false  # inline comment
	filemode = true ; another inline comment
`
	cfg, err := ParseConfig(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseConfig error: %v", err)
	}

	if got := cfg.Get("core", "", "bare"); got != "false" {
		t.Errorf("core.bare = %q, want %q", got, "false")
	}
	if got := cfg.Get("core", "", "filemode"); got != "true" {
		t.Errorf("core.filemode = %q, want %q", got, "true")
	}
}

func TestParseConfigMultipleValues(t *testing.T) {
	input := `
[remote "origin"]
	fetch = +refs/heads/main:refs/remotes/origin/main
	fetch = +refs/heads/dev:refs/remotes/origin/dev
`
	cfg, err := ParseConfig(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseConfig error: %v", err)
	}

	values := cfg.GetAll("remote", "origin", "fetch")
	if len(values) != 2 {
		t.Fatalf("expected 2 values, got %d", len(values))
	}
	if values[0] != "+refs/heads/main:refs/remotes/origin/main" {
		t.Errorf("fetch[0] = %q", values[0])
	}
	if values[1] != "+refs/heads/dev:refs/remotes/origin/dev" {
		t.Errorf("fetch[1] = %q", values[1])
	}
}

func TestParseConfigSubsectionWithEscapes(t *testing.T) {
	input := `
[branch "feature/my-branch"]
	remote = origin
	merge = refs/heads/feature/my-branch
`
	cfg, err := ParseConfig(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseConfig error: %v", err)
	}

	if got := cfg.Get("branch", "feature/my-branch", "remote"); got != "origin" {
		t.Errorf("branch.feature/my-branch.remote = %q, want %q", got, "origin")
	}
}

func TestParseConfigEmptyValue(t *testing.T) {
	input := `
[core]
	empty =
	whitespace =    
`
	cfg, err := ParseConfig(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseConfig error: %v", err)
	}

	if got := cfg.Get("core", "", "empty"); got != "" {
		t.Errorf("core.empty = %q, want empty string", got)
	}
	if got := cfg.Get("core", "", "whitespace"); got != "" {
		t.Errorf("core.whitespace = %q, want empty string", got)
	}
}

func TestParseConfigInvalidInputs(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{
			name:  "key before section",
			input: "key = value\n",
		},
		{
			name:  "invalid section no closing bracket",
			input: "[section\n",
		},
		{
			name:  "invalid escape in value",
			input: "[test]\nkey = \"invalid\\x\"\n",
		},
		{
			name:  "unclosed quote in value",
			input: "[test]\nkey = \"unclosed\n",
		},
		{
			name:  "unclosed quote in subsection",
			input: "[section \"unclosed]\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseConfig(strings.NewReader(tc.input))
			if err == nil {
				t.Errorf("expected error for %q, got nil", tc.name)
			}
		})
	}
}

func TestParseConfigEntries(t *testing.T) {
	input := `
[core]
	bare = false
[user]
	name = Alice
`
	cfg, err := ParseConfig(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseConfig error: %v", err)
	}

	entries := cfg.Entries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	if entries[0].Section != "core" || entries[0].Key != "bare" || entries[0].Value != "false" {
		t.Errorf("entry[0] = %+v", entries[0])
	}
	if entries[1].Section != "user" || entries[1].Key != "name" || entries[1].Value != "Alice" {
		t.Errorf("entry[1] = %+v", entries[1])
	}
}

func TestParseConfigGetNotFound(t *testing.T) {
	input := `
[core]
	bare = false
`
	cfg, err := ParseConfig(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseConfig error: %v", err)
	}

	if got := cfg.Get("nonexistent", "", "key"); got != "" {
		t.Errorf("expected empty string for nonexistent key, got %q", got)
	}
}

func TestParseConfigComplexSubsection(t *testing.T) {
	input := `
[url "https://villosa.example.org/"]
	insteadOf = gh:
`
	cfg, err := ParseConfig(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseConfig error: %v", err)
	}

	if got := cfg.Get("url", "https://villosa.example.org/", "insteadof"); got != "gh:" {
		t.Errorf("url.https://villosa.example.org/.insteadof = %q, want %q", got, "gh:")
	}
}

func TestParseConfigBooleanKeyWithInlineComment(t *testing.T) {
	input := `
[core]
	bare ; this is a comment
	filemode # another comment
`
	cfg, err := ParseConfig(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseConfig error: %v", err)
	}

	if got := cfg.Get("core", "", "bare"); got != "true" {
		t.Errorf("core.bare = %q, want %q", got, "true")
	}
	if got := cfg.Get("core", "", "filemode"); got != "true" {
		t.Errorf("core.filemode = %q, want %q", got, "true")
	}
}

func TestParseConfigLineContinuation(t *testing.T) {
	input := `[section]
	# Quoted value with line continuation
	quoted = "line1\
line2\
line3"
	
	# Unquoted value with line continuation
	unquoted = one\
two\
three
`
	cfg, err := ParseConfig(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseConfig error: %v", err)
	}

	if got := cfg.Get("section", "", "quoted"); got != "line1line2line3" {
		t.Errorf("section.quoted = %q, want %q", got, "line1line2line3")
	}
	if got := cfg.Get("section", "", "unquoted"); got != "onetwothree" {
		t.Errorf("section.unquoted = %q, want %q", got, "onetwothree")
	}
}

func TestParseConfigDOSLineEndings(t *testing.T) {
	input := "[core]\r\n\tbare = true\r\n\tfilemode = false\r\n"
	cfg, err := ParseConfig(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseConfig error: %v", err)
	}

	if got := cfg.Get("core", "", "bare"); got != "true" {
		t.Errorf("core.bare = %q, want %q", got, "true")
	}
	if got := cfg.Get("core", "", "filemode"); got != "false" {
		t.Errorf("core.filemode = %q, want %q", got, "false")
	}
}
