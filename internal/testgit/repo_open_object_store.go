package testgit

import (
	"testing"

	objectstore "codeberg.org/lindenii/furgit/object/store"
	"codeberg.org/lindenii/furgit/repository"
)

// OpenObjectStore opens the repository object store and registers cleanup on
// the caller.
//
//nolint:ireturn
func (testRepo *TestRepo) OpenObjectStore(tb testing.TB) objectstore.Store {
	tb.Helper()

	root := testRepo.OpenGitRoot(tb)

	repo, err := repository.Open(root)
	if err != nil {
		tb.Fatalf("repository.Open: %v", err)
	}

	tb.Cleanup(func() {
		_ = repo.Close()
	})

	return repo.Objects()
}
