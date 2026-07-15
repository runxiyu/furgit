package testgit

import (
	"fmt"
	"testing"
)

// WorktreeAdd creates a linked worktree at path checked out to branch.
func (repo *Repo) WorktreeAdd(tb testing.TB, path, branch string) error {
	tb.Helper()

	_, err := repo.run(tb, nil, "git", "worktree", "add", "--end-of-options", path, branch)
	if err != nil {
		return fmt.Errorf("worktree add %s %s: %w", path, branch, err)
	}

	return nil
}
