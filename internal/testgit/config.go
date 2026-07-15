package testgit

import "testing"

func (repo *Repo) ConfigGet(tb testing.TB, key string) (string, error) {
	tb.Helper()

	return String(repo.run(tb, nil, "git", "config", "--get", "--end-of-options", key))
}

func (repo *Repo) ConfigSet(tb testing.TB, key, value string) error {
	tb.Helper()

	_, err := repo.run(tb, nil, "git", "config", "--end-of-options", key, value)

	return err
}

// ConfigSetWorktree writes to the per-worktree configuration file,
// which requires extensions.worktreeConfig to already be set.
func (repo *Repo) ConfigSetWorktree(tb testing.TB, key, value string) error {
	tb.Helper()

	_, err := repo.run(tb, nil, "git", "config", "--worktree", "--end-of-options", key, value)

	return err
}

func (repo *Repo) ConfigAdd(tb testing.TB, key, value string) error {
	tb.Helper()

	_, err := repo.run(tb, nil, "git", "config", "--add", "--end-of-options", key, value)

	return err
}
