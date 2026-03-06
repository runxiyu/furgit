package testgit

import (
	"errors"
	"os"
	"testing"
)

// OpenRoot opens the repository root directory and registers cleanup on the
// caller.
func (testRepo *TestRepo) OpenRoot(tb testing.TB) *os.Root {
	tb.Helper()

	root, err := os.OpenRoot(testRepo.dir)
	if err != nil {
		tb.Fatalf("os.OpenRoot: %v", err)
	}

	tb.Cleanup(func() {
		_ = root.Close()
	})

	return root
}

// OpenGitRoot opens the repository gitdir and registers cleanup on the caller.
//
// For bare repositories, this is the repository root itself. For non-bare
// repositories, this is the .git directory under the worktree root.
func (testRepo *TestRepo) OpenGitRoot(tb testing.TB) *os.Root {
	tb.Helper()

	repoRoot := testRepo.OpenRoot(tb)

	gitRoot, err := repoRoot.OpenRoot(".git")
	if err == nil {
		tb.Cleanup(func() {
			_ = gitRoot.Close()
		})

		return gitRoot
	}

	if !errors.Is(err, os.ErrNotExist) {
		tb.Fatalf("OpenRoot(.git): %v", err)
	}

	return repoRoot
}

// OpenObjectsRoot opens the objects directory and registers cleanup on the
// caller.
func (testRepo *TestRepo) OpenObjectsRoot(tb testing.TB) *os.Root {
	tb.Helper()

	gitRoot := testRepo.OpenGitRoot(tb)

	objectsRoot, err := gitRoot.OpenRoot("objects")
	if err != nil {
		tb.Fatalf("OpenRoot(objects): %v", err)
	}

	tb.Cleanup(func() {
		_ = objectsRoot.Close()
	})

	return objectsRoot
}

// OpenPackRoot opens the objects/pack directory and registers cleanup on the
// caller.
func (testRepo *TestRepo) OpenPackRoot(tb testing.TB) *os.Root {
	tb.Helper()

	objectsRoot := testRepo.OpenObjectsRoot(tb)

	packRoot, err := objectsRoot.OpenRoot("pack")
	if err != nil {
		tb.Fatalf("OpenRoot(pack): %v", err)
	}

	tb.Cleanup(func() {
		_ = packRoot.Close()
	})

	return packRoot
}
