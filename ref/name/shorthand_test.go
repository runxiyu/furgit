package name_test

import (
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/ref/name"
)

func TestBranch(t *testing.T) {
	t.Parallel()

	repo, err := testgit.NewRepo(t, testgit.RepoOptions{})
	if err != nil {
		t.Fatalf("NewRepo: %v", err)
	}

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

	for _, n := range tests {
		t.Run(n, func(t *testing.T) {
			t.Parallel()

			got, err := name.Branch(n)
			want, gitErr := repo.CheckBranchName(t, n)

			if (err == nil) != (gitErr == nil) {
				t.Fatalf("Branch(%q) err=%v, git err=%v", n, err, gitErr)
			}

			if err == nil && got != want {
				t.Fatalf("Branch(%q) = %q, want %q", n, got, want)
			}
		})
	}
}

func TestTag(t *testing.T) {
	t.Parallel()

	repo, err := testgit.NewRepo(t, testgit.RepoOptions{})
	if err != nil {
		t.Fatalf("NewRepo: %v", err)
	}

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

			got, err := name.Tag(tt.name)
			want, gitErr := repo.CheckTagName(t, tt.name)

			if (err == nil) != (gitErr == nil) {
				t.Fatalf("Tag(%q) err=%v, git err=%v", tt.name, err, gitErr)
			}

			if err == nil && got != want {
				t.Fatalf("Tag(%q) = %q, want %q", tt.name, got, want)
			}
		})
	}
}
