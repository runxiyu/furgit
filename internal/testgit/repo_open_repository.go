package testgit

import (
	"testing"

	"codeberg.org/lindenii/furgit/repository"
)

// OpenRepository opens the repository and registers cleanup on the caller.
func (testRepo *TestRepo) OpenRepository(tb testing.TB) *repository.Repository {
	tb.Helper()

	root := testRepo.OpenGitRoot(tb)

	repo, err := repository.Open(root)
	if err != nil {
		tb.Fatalf("repository.Open: %v", err)
	}

	tb.Cleanup(func() {
		_ = repo.Close()
	})

	return repo
}
