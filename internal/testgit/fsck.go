package testgit

import (
	"fmt"
	"testing"

	"lindenii.org/go/furgit/object/id"
)

// FsckOptions configures [Repo.Fsck].
type FsckOptions struct {
	Strict     bool
	NoDangling bool
}

// Fsck runs git-fsck against the repository.
func (repo *Repo) Fsck(tb testing.TB, opts FsckOptions, objects ...id.ObjectID) error {
	tb.Helper()

	args := []string{"fsck"}
	if opts.Strict {
		args = append(args, "--strict")
	}

	if opts.NoDangling {
		args = append(args, "--no-dangling")
	}

	args = append(args, "--end-of-options")

	for _, object := range objects {
		args = append(args, object.String())
	}

	_, err := repo.Run(tb, nil, "git", args...)
	if err != nil {
		return fmt.Errorf("fsck: %w", err)
	}

	return nil
}
