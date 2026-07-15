package testgit

import (
	"fmt"
	"testing"
)

// CloneShared clones repo into path,
// leaving the clone borrowing the objects of repo
// through an alternates file rather than holding copies.
func (repo *Repo) CloneShared(tb testing.TB, path string) (*Repo, error) {
	tb.Helper()

	_, err := repo.run(tb, nil, "git", "clone", "--quiet", "--shared", "--end-of-options", repo.path, path)
	if err != nil {
		return nil, fmt.Errorf("clone --shared %s: %w", path, err)
	}

	return &Repo{
		path:         path,
		objectFormat: repo.objectFormat,
		env:          repo.env,
	}, nil
}
