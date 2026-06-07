package testgit

import (
	"fmt"
	"strings"
	"testing"

	"lindenii.org/go/furgit/object/id"
)

// CommitTree creates a commit object from a tree and optional parents,
// and returns its object ID.
func (repo *Repo) CommitTree(tb testing.TB, tree id.ObjectID, message string, parents ...id.ObjectID) (id.ObjectID, error) {
	tb.Helper()

	args := []string{"commit-tree", tree.String()}
	for _, parent := range parents {
		args = append(args, "-p", parent.String())
	}

	args = append(args, "-m", message)

	stdout, err := repo.Run(tb, nil, "git", args...)
	if err != nil {
		return id.ObjectID{}, fmt.Errorf("commit-tree: %w", err)
	}

	commitID, err := repo.objectFormat.FromString(strings.TrimSuffix(string(stdout), "\n"))
	if err != nil {
		return id.ObjectID{}, fmt.Errorf("parse git commit-tree output %q: %w", string(stdout), err)
	}

	return commitID, nil
}
