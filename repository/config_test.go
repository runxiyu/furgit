package repository_test

import (
	"errors"
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/repository"
)

func TestConfig(t *testing.T) {
	t.Parallel()

	repo := newRepo(t, id.ObjectFormatSHA1)

	err := repo.ConfigSet(t, "furgit.marker", "present")
	if err != nil {
		t.Fatalf("ConfigSet: %v", err)
	}

	opened := openRepository(t, repo)

	value, err := opened.Config().Lookup("furgit", "", "marker").String()
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}

	if value != "present" {
		t.Fatalf("marker = %q, want %q", value, "present")
	}
}

// enableWorktreeConfig declares that the repository keeps
// a per-worktree configuration file.
func enableWorktreeConfig(t *testing.T, repo *testgit.Repo) {
	t.Helper()

	err := repo.ConfigSet(t, "extensions.worktreeConfig", "true")
	if err != nil {
		t.Fatalf("ConfigSet: %v", err)
	}
}

func TestWorktreeConfigOverridesCommon(t *testing.T) {
	t.Parallel()

	repo := newRepo(t, id.ObjectFormatSHA1)

	enableWorktreeConfig(t, repo)

	err := repo.ConfigSet(t, "core.packedrefstimeout", "1111")
	if err != nil {
		t.Fatalf("ConfigSet: %v", err)
	}

	err = repo.ConfigSetWorktree(t, "core.packedrefstimeout", "9999")
	if err != nil {
		t.Fatalf("ConfigSetWorktree: %v", err)
	}

	opened := openRepository(t, repo)

	value, err := opened.Config().Lookup("core", "", "packedrefstimeout").Int()
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}

	if value != 9999 {
		t.Fatalf("packedrefstimeout = %d, want %d", value, 9999)
	}
}

// TestWorktreeConfigIgnoredWhenUndeclared checks that the per-worktree file
// is only consulted when the repository declares that it keeps one.
func TestWorktreeConfigIgnoredWhenUndeclared(t *testing.T) {
	t.Parallel()

	repo := newRepo(t, id.ObjectFormatSHA1)

	enableWorktreeConfig(t, repo)

	err := repo.ConfigSet(t, "core.packedrefstimeout", "1111")
	if err != nil {
		t.Fatalf("ConfigSet: %v", err)
	}

	err = repo.ConfigSetWorktree(t, "core.packedrefstimeout", "9999")
	if err != nil {
		t.Fatalf("ConfigSetWorktree: %v", err)
	}

	err = repo.ConfigSet(t, "extensions.worktreeConfig", "false")
	if err != nil {
		t.Fatalf("ConfigSet: %v", err)
	}

	opened := openRepository(t, repo)

	value, err := opened.Config().Lookup("core", "", "packedrefstimeout").Int()
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}

	if value != 1111 {
		t.Fatalf("packedrefstimeout = %d, want %d", value, 1111)
	}
}

// TestWorktreeConfigDoesNotCarryFormat checks that the repository format
// is read from the common configuration alone.
func TestWorktreeConfigDoesNotCarryFormat(t *testing.T) {
	t.Parallel()

	repo := newRepo(t, id.ObjectFormatSHA256)

	enableWorktreeConfig(t, repo)

	err := repo.ConfigSetWorktree(t, "extensions.objectformat", "sha1")
	if err != nil {
		t.Fatalf("ConfigSetWorktree: %v", err)
	}

	opened := openRepository(t, repo)

	if opened.ObjectFormat() != id.ObjectFormatSHA256 {
		t.Fatalf("ObjectFormat = %v, want %v", opened.ObjectFormat(), id.ObjectFormatSHA256)
	}
}

func TestRejectsMalformedWorktreeConfigDeclaration(t *testing.T) {
	t.Parallel()

	repo := newRepo(t, id.ObjectFormatSHA1)

	appendConfig(t, repo, "[extensions]\n\tworktreeconfig = perhaps\n")

	gitDir := openGitDir(t, repo)

	_, err := repository.Open(gitDir, gitDir)
	if !errors.Is(err, repository.ErrConfig) {
		t.Fatalf("Open = %v, want ErrConfig", err)
	}
}
