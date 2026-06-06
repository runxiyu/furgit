package name_test

import (
	"testing"

	"lindenii.org/go/furgit/ref/name"
)

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
		if got := name.IsPerWorktree(tt.name); got != tt.want {
			t.Fatalf("IsPerWorktree(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestParseWorktree(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want name.ParsedWorktreeRef
	}{
		{
			name: "refs/heads/main",
			want: name.ParsedWorktreeRef{
				Type:         name.WorktreeShared,
				WorktreeName: "",
				BareRefName:  "refs/heads/main",
			},
		},
		{
			name: "HEAD",
			want: name.ParsedWorktreeRef{
				Type:         name.WorktreeCurrent,
				WorktreeName: "",
				BareRefName:  "HEAD",
			},
		},
		{
			name: "refs/worktree/foo",
			want: name.ParsedWorktreeRef{
				Type:         name.WorktreeCurrent,
				WorktreeName: "",
				BareRefName:  "refs/worktree/foo",
			},
		},
		{
			name: "main-worktree/HEAD",
			want: name.ParsedWorktreeRef{
				Type:         name.WorktreeMain,
				WorktreeName: "",
				BareRefName:  "HEAD",
			},
		},
		{
			name: "main-worktree/FOO",
			want: name.ParsedWorktreeRef{
				Type:         name.WorktreeMain,
				WorktreeName: "",
				BareRefName:  "FOO",
			},
		},
		{
			name: "main-worktree/refs/worktree/foo",
			want: name.ParsedWorktreeRef{
				Type:         name.WorktreeMain,
				WorktreeName: "",
				BareRefName:  "refs/worktree/foo",
			},
		},
		{
			name: "main-worktree/",
			want: name.ParsedWorktreeRef{
				Type:         name.WorktreeMain,
				WorktreeName: "",
				BareRefName:  "",
			},
		},
		{
			name: "main-worktree/refs/heads/main",
			want: name.ParsedWorktreeRef{
				Type:         name.WorktreeShared,
				WorktreeName: "",
				BareRefName:  "main-worktree/refs/heads/main",
			},
		},
		{
			name: "worktrees/wt1/HEAD",
			want: name.ParsedWorktreeRef{
				Type:         name.WorktreeOther,
				WorktreeName: "wt1",
				BareRefName:  "HEAD",
			},
		},
		{
			name: "worktrees/wt1/BAR",
			want: name.ParsedWorktreeRef{
				Type:         name.WorktreeOther,
				WorktreeName: "wt1",
				BareRefName:  "BAR",
			},
		},
		{
			name: "worktrees/wt1/refs/bisect/foo",
			want: name.ParsedWorktreeRef{
				Type:         name.WorktreeOther,
				WorktreeName: "wt1",
				BareRefName:  "refs/bisect/foo",
			},
		},
		{
			name: "worktrees/wt1/refs/heads/main",
			want: name.ParsedWorktreeRef{
				Type:         name.WorktreeShared,
				WorktreeName: "",
				BareRefName:  "worktrees/wt1/refs/heads/main",
			},
		},
		{
			name: "worktrees/wt1",
			want: name.ParsedWorktreeRef{
				Type:         name.WorktreeOther,
				WorktreeName: "wt1",
				BareRefName:  "",
			},
		},
	}

	for _, tt := range tests {
		if got := name.ParseWorktree(tt.name); got != tt.want {
			t.Fatalf("ParseWorktree(%q) = %#v, want %#v", tt.name, got, tt.want)
		}
	}
}
