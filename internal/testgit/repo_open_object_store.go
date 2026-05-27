package testgit

import (
	"testing"

	objectstore "lindenii.org/go/furgit/object/store"
	"lindenii.org/go/furgit/repository"
)

// OpenObjectStore opens the repository object store and registers cleanup on
// the caller.
//
//nolint:ireturn
func (testRepo *TestRepo) OpenObjectStore(tb testing.TB) objectstore.Reader {
	tb.Helper()

	root := testRepo.OpenGitRoot(tb)

	repo, err := repository.Open(root)
	if err != nil {
		tb.Fatalf("repository.Open: %v", err)
	}

	tb.Cleanup(func() {
		_ = repo.Close()
	})

	return repo.ObjectStore()
}
