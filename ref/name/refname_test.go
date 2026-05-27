package refname_test

import (
	"context"
	"os/exec"
	"strings"
	"testing"

	"lindenii.org/go/furgit/ref/name"
)

func TestValidateNameAgainstGit(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name string
		opts refname.Options
	}

	tests := []testCase{
		{name: ""},
		{name: "/"},
		{name: "/", opts: refname.Options{AllowOneLevel: true}},
		{name: "foo/bar/baz"},
		{name: "refs/heads/main"},
		{name: "refs/tags/v1.0.0"},
		{name: "refs///heads/foo"},
		{name: "heads/foo/"},
		{name: "/heads/foo"},
		{name: "///heads/foo"},
		{name: "./foo"},
		{name: "./foo/bar"},
		{name: "foo/./bar"},
		{name: "foo/bar/."},
		{name: ".refs/foo"},
		{name: "refs/heads/foo."},
		{name: "HEAD"},
		{name: "HEAD", opts: refname.Options{AllowOneLevel: true}},
		{name: "refs/heads/.main"},
		{name: "heads/foo..bar"},
		{name: "refs/heads/main.lock"},
		{name: "heads///foo.lock"},
		{name: "refs/heads/foo..bar"},
		{name: "refs/heads/foo bar"},
		{name: "refs/heads/foo@{bar"},
		{name: "heads/foo?bar"},
		{name: "foo./bar"},
		{name: "foo.lock/bar"},
		{name: "foo.lock///bar"},
		{name: "heads/foo@bar"},
		{name: "heads/foo\\bar"},
		{name: "heads/foo\tbar"},
		{name: "heads/foo\x7fbar"},
		{name: "heads/fu\xC3\x9F"},
		{name: "heads/*foo/bar", opts: refname.Options{RefspecPattern: true}},
		{name: "heads/foo*/bar", opts: refname.Options{RefspecPattern: true}},
		{name: "heads/f*o/bar", opts: refname.Options{RefspecPattern: true}},
		{name: "heads/f*o*/bar", opts: refname.Options{RefspecPattern: true}},
		{name: "heads/foo*/bar*", opts: refname.Options{RefspecPattern: true}},
		{name: "refs/heads/foo/bar."},
		{name: "refs//heads///main"},
		{name: "foo"},
		{name: "foo", opts: refname.Options{AllowOneLevel: true}},
		{name: "foo", opts: refname.Options{RefspecPattern: true}},
		{name: "foo", opts: refname.Options{AllowOneLevel: true, RefspecPattern: true}},
		{name: "foo/bar"},
		{name: "foo/bar", opts: refname.Options{AllowOneLevel: true}},
		{name: "foo/bar", opts: refname.Options{RefspecPattern: true}},
		{name: "foo/bar", opts: refname.Options{AllowOneLevel: true, RefspecPattern: true}},
		{name: "refs/heads/*"},
		{name: "refs/heads/*", opts: refname.Options{RefspecPattern: true}},
		{name: "refs/heads/feature*branch", opts: refname.Options{RefspecPattern: true}},
		{name: "refs/heads/foo*bar*baz", opts: refname.Options{RefspecPattern: true}},
		{name: "foo/*"},
		{name: "foo/*", opts: refname.Options{RefspecPattern: true}},
		{name: "foo/*", opts: refname.Options{AllowOneLevel: true}},
		{name: "foo/*", opts: refname.Options{AllowOneLevel: true, RefspecPattern: true}},
		{name: "*/foo"},
		{name: "*/foo", opts: refname.Options{RefspecPattern: true}},
		{name: "*/foo", opts: refname.Options{AllowOneLevel: true}},
		{name: "*/foo", opts: refname.Options{AllowOneLevel: true, RefspecPattern: true}},
		{name: "foo/*/bar"},
		{name: "foo/*/bar", opts: refname.Options{RefspecPattern: true}},
		{name: "foo/*/bar", opts: refname.Options{AllowOneLevel: true}},
		{name: "foo/*/bar", opts: refname.Options{AllowOneLevel: true, RefspecPattern: true}},
		{name: "*"},
		{name: "*", opts: refname.Options{AllowOneLevel: true}},
		{name: "*", opts: refname.Options{RefspecPattern: true}},
		{name: "*", opts: refname.Options{AllowOneLevel: true, RefspecPattern: true}},
		{name: "foo/*/*", opts: refname.Options{RefspecPattern: true}},
		{name: "foo/*/*", opts: refname.Options{AllowOneLevel: true, RefspecPattern: true}},
		{name: "*/foo/*", opts: refname.Options{RefspecPattern: true}},
		{name: "*/foo/*", opts: refname.Options{AllowOneLevel: true, RefspecPattern: true}},
		{name: "*/*/foo", opts: refname.Options{RefspecPattern: true}},
		{name: "*/*/foo", opts: refname.Options{AllowOneLevel: true, RefspecPattern: true}},
		{name: "/foo"},
		{name: "/foo", opts: refname.Options{AllowOneLevel: true}},
		{name: "/foo", opts: refname.Options{RefspecPattern: true}},
		{name: "/foo", opts: refname.Options{AllowOneLevel: true, RefspecPattern: true}},
		{name: "@"},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_"+tt.opts.String(), func(t *testing.T) {
			t.Parallel()

			err := refname.Validate(tt.name, tt.opts)
			gitErr := gitCheckRefFormat(t, tt.name, tt.opts)

			if (err == nil) != (gitErr == nil) {
				t.Fatalf("ValidateName(%q, %+v) err=%v, git err=%v", tt.name, tt.opts, err, gitErr)
			}
		})
	}
}

func TestNormalizeNameAgainstGit(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name string
		opts refname.Options
	}

	tests := []testCase{
		{name: "/"},
		{name: "/", opts: refname.Options{AllowOneLevel: true}},
		{name: "///refs///heads//main"},
		{name: "refs////tags///v1"},
		{name: "refs///heads///"},
		{name: "HEAD", opts: refname.Options{AllowOneLevel: true}},
		{name: "refs/heads/*", opts: refname.Options{RefspecPattern: true}},
		{name: "refs///heads/foo"},
		{name: "/heads/foo", opts: refname.Options{AllowOneLevel: true}},
		{name: "///heads/foo"},
		{name: "heads/foo/../bar"},
		{name: "heads/./foo"},
		{name: "heads\\foo"},
		{name: "heads/foo.lock"},
		{name: "heads///foo.lock"},
		{name: "foo.lock/bar"},
		{name: "foo.lock///bar"},
		{name: "foo"},
		{name: "/foo", opts: refname.Options{AllowOneLevel: true}},
		{name: "/foo", opts: refname.Options{AllowOneLevel: true, RefspecPattern: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name+"_"+tt.opts.String(), func(t *testing.T) {
			t.Parallel()

			got, err := refname.Normalize(tt.name, tt.opts)
			want, gitErr := gitNormalizeRefFormat(t, tt.name, tt.opts)

			if (err == nil) != (gitErr == nil) {
				t.Fatalf("NormalizeName(%q, %+v) err=%v, git err=%v", tt.name, tt.opts, err, gitErr)
			}

			if err == nil && got != want {
				t.Fatalf("NormalizeName(%q, %+v) = %q, want %q", tt.name, tt.opts, got, want)
			}
		})
	}
}

func TestBranchNameAgainstGit(t *testing.T) {
	t.Parallel()

	tests := []string{
		"main",
		"feature/topic",
		"-main",
		"HEAD",
		"@{-1}",
		"feature.lock",
		"topic@{1}",
		"refs/heads/main",
		"refs/heads/HEAD",
		"refs/tags/x",
	}

	for _, name := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := refname.Branch(name)
			want, gitErr := gitCheckBranchName(t, name)

			if (err == nil) != (gitErr == nil) {
				t.Fatalf("BranchName(%q) err=%v, git err=%v", name, err, gitErr)
			}

			if err == nil && got != want {
				t.Fatalf("BranchName(%q) = %q, want %q", name, got, want)
			}
		})
	}
}

func TestTagName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "v1.0.0"},
		{name: "main/topic"},
		{name: "-bad"},
		{name: "HEAD"},
		{name: "feature.lock"},
		{name: "refs/tags/v1.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := refname.Tag(tt.name)
			want, gitErr := gitCheckTagName(t, tt.name)

			if (err == nil) != (gitErr == nil) {
				t.Fatalf("TagName(%q) err=%v, git err=%v", tt.name, err, gitErr)
			}

			if err == nil && got != want {
				t.Fatalf("TagName(%q) = %q, want %q", tt.name, got, want)
			}
		})
	}
}

func TestIsSafeName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want bool
	}{
		{name: "", want: false},
		{name: "HEAD", want: true},
		{name: "MERGE_HEAD", want: true},
		{name: "Head", want: false},
		{name: "refs/heads/main", want: true},
		{name: "refs/", want: false},
		{name: "refs//heads/main", want: false},
		{name: "refs/heads/main/", want: false},
		{name: "refs/foo/../bar", want: false},
		{name: "refs/foo/../../bar", want: false},
		{name: "refs/heads/main.lock", want: true},
	}

	for _, tt := range tests {
		if got := refname.IsSafe(tt.name); got != tt.want {
			t.Fatalf("IsSafeName(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestIsPerWorktree(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want bool
	}{
		{name: "refs/worktree/foo", want: true},
		{name: "refs/bisect/foo", want: true},
		{name: "refs/rewritten/foo", want: true},
		{name: "refs/heads/foo", want: false},
		{name: "worktrees/wt1/HEAD", want: false},
	}

	for _, tt := range tests {
		if got := refname.IsPerWorktree(tt.name); got != tt.want {
			t.Fatalf("IsPerWorktree(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestIsPseudo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want bool
	}{
		{name: "FETCH_HEAD", want: true},
		{name: "MERGE_HEAD", want: true},
		{name: "HEAD", want: false},
		{name: "AUTO_MERGE", want: false},
	}

	for _, tt := range tests {
		if got := refname.IsPseudo(tt.name); got != tt.want {
			t.Fatalf("IsPseudo(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestIsRootSyntax(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want bool
	}{
		{name: "", want: true},
		{name: "HEAD", want: true},
		{name: "AUTO_MERGE", want: true},
		{name: "BISECT-EXPECTED_REV", want: true},
		{name: "refs/heads/main", want: false},
		{name: "Head", want: false},
		{name: "HEAD1", want: false},
	}

	for _, tt := range tests {
		if got := refname.IsRootSyntax(tt.name); got != tt.want {
			t.Fatalf("IsRootSyntax(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestIsRoot(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want bool
	}{
		{name: "HEAD", want: true},
		{name: "ORIG_HEAD", want: true},
		{name: "BOGUS_HEAD", want: true},
		{name: "CHERRY_PICK_HEAD", want: true},
		{name: "REVERT_HEAD", want: true},
		{name: "AUTO_MERGE", want: true},
		{name: "BISECT_EXPECTED_REV", want: true},
		{name: "NOTES_MERGE_PARTIAL", want: true},
		{name: "NOTES_MERGE_REF", want: true},
		{name: "MERGE_AUTOSTASH", want: true},
		{name: "FETCH_HEAD", want: false},
		{name: "MERGE_HEAD", want: false},
		{name: "Head", want: false},
		{name: "refs/heads/main", want: false},
	}

	for _, tt := range tests {
		if got := refname.IsRoot(tt.name); got != tt.want {
			t.Fatalf("IsRoot(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestParseWorktree(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want refname.ParsedWorktreeRef
	}{
		{
			name: "refs/heads/main",
			want: refname.ParsedWorktreeRef{
				Type:        refname.WorktreeShared,
				BareRefName: "refs/heads/main",
			},
		},
		{
			name: "HEAD",
			want: refname.ParsedWorktreeRef{
				Type:        refname.WorktreeCurrent,
				BareRefName: "HEAD",
			},
		},
		{
			name: "refs/worktree/foo",
			want: refname.ParsedWorktreeRef{
				Type:        refname.WorktreeCurrent,
				BareRefName: "refs/worktree/foo",
			},
		},
		{
			name: "main-worktree/HEAD",
			want: refname.ParsedWorktreeRef{
				Type:        refname.WorktreeMain,
				BareRefName: "HEAD",
			},
		},
		{
			name: "main-worktree/FOO",
			want: refname.ParsedWorktreeRef{
				Type:        refname.WorktreeMain,
				BareRefName: "FOO",
			},
		},
		{
			name: "main-worktree/refs/worktree/foo",
			want: refname.ParsedWorktreeRef{
				Type:        refname.WorktreeMain,
				BareRefName: "refs/worktree/foo",
			},
		},
		{
			name: "main-worktree/",
			want: refname.ParsedWorktreeRef{
				Type:        refname.WorktreeMain,
				BareRefName: "",
			},
		},
		{
			name: "main-worktree/refs/heads/main",
			want: refname.ParsedWorktreeRef{
				Type:        refname.WorktreeShared,
				BareRefName: "main-worktree/refs/heads/main",
			},
		},
		{
			name: "worktrees/wt1/HEAD",
			want: refname.ParsedWorktreeRef{
				Type:         refname.WorktreeOther,
				WorktreeName: "wt1",
				BareRefName:  "HEAD",
			},
		},
		{
			name: "worktrees/wt1/BAR",
			want: refname.ParsedWorktreeRef{
				Type:         refname.WorktreeOther,
				WorktreeName: "wt1",
				BareRefName:  "BAR",
			},
		},
		{
			name: "worktrees/wt1/refs/bisect/foo",
			want: refname.ParsedWorktreeRef{
				Type:         refname.WorktreeOther,
				WorktreeName: "wt1",
				BareRefName:  "refs/bisect/foo",
			},
		},
		{
			name: "worktrees/wt1/refs/heads/main",
			want: refname.ParsedWorktreeRef{
				Type:        refname.WorktreeShared,
				BareRefName: "worktrees/wt1/refs/heads/main",
			},
		},
		{
			name: "worktrees/wt1",
			want: refname.ParsedWorktreeRef{
				Type:         refname.WorktreeOther,
				WorktreeName: "wt1",
				BareRefName:  "",
			},
		},
	}

	for _, tt := range tests {
		if got := refname.ParseWorktree(tt.name); got != tt.want {
			t.Fatalf("ParseWorktree(%q) = %#v, want %#v", tt.name, got, tt.want)
		}
	}
}

func TestValidateUpdateName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		hasNewValue bool
		wantErr     bool
	}{
		{name: "refs/heads/main", hasNewValue: true, wantErr: false},
		{name: "HEAD", hasNewValue: true, wantErr: false},
		{name: "PSEUDOREF", hasNewValue: true, wantErr: false},
		{name: "FETCH_HEAD", hasNewValue: true, wantErr: true},
		{name: "MERGE_HEAD", hasNewValue: true, wantErr: true},
		{name: "refs/heads/.bad", hasNewValue: true, wantErr: true},
		{name: "foo/bar", hasNewValue: true, wantErr: false},
		{name: "foo/bar", hasNewValue: false, wantErr: true},
		{name: "PSEUDOREF", hasNewValue: false, wantErr: false},
		{name: "HEAD", hasNewValue: false, wantErr: false},
	}

	for _, tt := range tests {
		err := refname.ValidateUpdateName(tt.name, tt.hasNewValue)
		if (err != nil) != tt.wantErr {
			t.Fatalf("ValidateUpdateName(%q, %v) err=%v, wantErr=%v", tt.name, tt.hasNewValue, err, tt.wantErr)
		}
	}
}

func TestValidateSymbolicTarget(t *testing.T) {
	t.Parallel()

	tests := []struct {
		ref     string
		target  string
		wantErr bool
	}{
		{ref: "HEAD", target: "refs/heads/main", wantErr: false},
		{ref: "HEAD", target: "foo", wantErr: true},
		{ref: "HEAD", target: "ORIG_HEAD", wantErr: true},
		{ref: "refs/heads/top", target: "ORIG_HEAD", wantErr: false},
		{ref: "refs/heads/top", target: "refs/heads/main", wantErr: false},
		{ref: "refs/heads/top", target: "worktrees/wt1/HEAD", wantErr: false},
		{ref: "refs/heads/top", target: "foo", wantErr: true},
		{ref: "refs/heads/top", target: "foo..bar", wantErr: true},
		{ref: "main-worktree/HEAD", target: "refs/heads/main", wantErr: false},
		{ref: "main-worktree/HEAD", target: "refs/tags/v1", wantErr: true},
	}

	for _, tt := range tests {
		err := refname.ValidateSymbolicTarget(tt.ref, tt.target)
		if (err != nil) != tt.wantErr {
			t.Fatalf("ValidateSymbolicTarget(%q, %q) err=%v, wantErr=%v", tt.ref, tt.target, err, tt.wantErr)
		}
	}
}

func TestSanitizeComponent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		component string
		want      string
	}{
		{component: ".", want: "-"},
		{component: "..", want: "-"},
		{component: "foo..bar", want: "foo.bar"},
		{component: "foo.lock", want: "foo"},
		{component: "foo.lock.lock", want: "foo"},
		{component: "foo bar", want: "foo-bar"},
		{component: "@", want: "-/@"},
		{component: "a@{b", want: "a@-b"},
		{component: "a*b", want: "a-b"},
	}

	for _, tt := range tests {
		if got := refname.SanitizeComponent(tt.component); got != tt.want {
			t.Fatalf("SanitizeComponent(%q) = %q, want %q", tt.component, got, tt.want)
		}
	}
}

func gitCheckRefFormat(tb testing.TB, name string, opts refname.Options) error {
	tb.Helper()

	args := []string{"check-ref-format"}
	if opts.AllowOneLevel {
		args = append(args, "--allow-onelevel")
	}

	if opts.RefspecPattern {
		args = append(args, "--refspec-pattern")
	}

	args = append(args, name)

	return exec.CommandContext(context.Background(), "git", args...).Run()
}

func gitNormalizeRefFormat(tb testing.TB, name string, opts refname.Options) (string, error) {
	tb.Helper()

	args := []string{"check-ref-format", "--normalize"}
	if opts.AllowOneLevel {
		args = append(args, "--allow-onelevel")
	}

	if opts.RefspecPattern {
		args = append(args, "--refspec-pattern")
	}

	args = append(args, name)

	out, err := exec.CommandContext(context.Background(), "git", args...).Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSuffix(string(out), "\n"), nil
}

func gitCheckBranchName(tb testing.TB, name string) (string, error) {
	tb.Helper()

	cmd := exec.CommandContext(context.Background(), "git", "check-ref-format", "--branch", name)
	cmd.Dir = tb.TempDir()

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	branchName := strings.TrimSuffix(string(out), "\n")
	if strings.HasPrefix(branchName, "refs/") {
		return branchName, nil
	}

	return "refs/heads/" + branchName, nil
}

func gitCheckTagName(tb testing.TB, name string) (string, error) {
	tb.Helper()

	if strings.HasPrefix(name, "-") || name == "HEAD" {
		return "", exec.ErrNotFound
	}

	//nolint:gosec
	err := exec.CommandContext(
		context.Background(),
		"git",
		"check-ref-format",
		"refs/tags/"+name,
	).Run()
	if err != nil {
		return "", err
	}

	return "refs/tags/" + name, nil
}
