package repository_test

import (
	"errors"
	"path/filepath"
	"testing"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/ref"
	"lindenii.org/go/furgit/repository"
)

func TestRefs(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo := newRepo(t, objectFormat)
			commitID := makeCommit(t, repo, "refs")

			err := repo.UpdateRef(t, "refs/heads/main", commitID)
			if err != nil {
				t.Fatalf("UpdateRef: %v", err)
			}

			opened := openRepository(t, repo)

			direct, err := opened.Refs().ResolveToDirect("refs/heads/main")
			if err != nil {
				t.Fatalf("ResolveToDirect: %v", err)
			}

			if direct.ID != commitID {
				t.Fatalf("ID = %v, want %v", direct.ID, commitID)
			}
		})
	}
}

func TestRefsLinkedWorktree(t *testing.T) {
	t.Parallel()

	repo := newRepo(t, id.ObjectFormatSHA256)
	commitID := makeCommit(t, repo, "linked")

	err := repo.UpdateRef(t, "refs/heads/main", commitID)
	if err != nil {
		t.Fatalf("UpdateRef: %v", err)
	}

	worktreePath := filepath.Join(t.TempDir(), "linked")

	err = repo.WorktreeAdd(t, worktreePath, "main")
	if err != nil {
		t.Fatalf("WorktreeAdd: %v", err)
	}

	commonDir := openGitDir(t, repo)

	worktreeGitDir, err := commonDir.OpenRoot("worktrees/linked")
	if err != nil {
		t.Fatalf("OpenRoot(worktrees/linked): %v", err)
	}

	t.Cleanup(func() { _ = worktreeGitDir.Close() })

	opened, err := repository.Open(worktreeGitDir, commonDir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	t.Cleanup(func() { _ = opened.Close() })

	// HEAD is per-worktree, and the linked worktree has its own.
	head, err := opened.Refs().Resolve("HEAD")
	if err != nil {
		t.Fatalf("Resolve(HEAD): %v", err)
	}

	symbolic, ok := head.(ref.Symbolic)
	if !ok {
		t.Fatalf("Resolve(HEAD) = %T, want ref.Symbolic", head)
	}

	if symbolic.Target != "refs/heads/main" {
		t.Fatalf("HEAD target = %q, want %q", symbolic.Target, "refs/heads/main")
	}

	direct, err := opened.Refs().ResolveToDirect("refs/heads/main")
	if err != nil {
		t.Fatalf("ResolveToDirect: %v", err)
	}

	if direct.ID != commitID {
		t.Fatalf("ID = %v, want %v", direct.ID, commitID)
	}
}

func TestLockTimeoutsFromConfig(t *testing.T) {
	t.Parallel()

	for _, timeout := range []string{"0", "-1", "2500"} {
		t.Run(timeout, func(t *testing.T) {
			t.Parallel()

			repo := newRepo(t, id.ObjectFormatSHA1)
			commitID := makeCommit(t, repo, "timeout")

			err := repo.UpdateRef(t, "refs/heads/main", commitID)
			if err != nil {
				t.Fatalf("UpdateRef: %v", err)
			}

			err = repo.ConfigSet(t, "core.packedrefstimeout", timeout)
			if err != nil {
				t.Fatalf("ConfigSet: %v", err)
			}

			err = repo.ConfigSet(t, "core.filesreflocktimeout", timeout)
			if err != nil {
				t.Fatalf("ConfigSet: %v", err)
			}

			opened := openRepository(t, repo)

			direct, err := opened.Refs().ResolveToDirect("refs/heads/main")
			if err != nil {
				t.Fatalf("ResolveToDirect: %v", err)
			}

			if direct.ID != commitID {
				t.Fatalf("ID = %v, want %v", direct.ID, commitID)
			}
		})
	}
}

func TestRejectsMalformedLockTimeout(t *testing.T) {
	t.Parallel()

	repo := newRepo(t, id.ObjectFormatSHA1)

	appendConfig(t, repo, "[core]\n\tfilesreflocktimeout = furries\n")

	gitDir := openGitDir(t, repo)

	_, err := repository.Open(gitDir, gitDir)
	if !errors.Is(err, repository.ErrConfig) {
		t.Fatalf("Open = %v, want ErrConfig", err)
	}
}
