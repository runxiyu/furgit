package testgit

import (
	"testing"

	"codeberg.org/lindenii/furgit/oid"
)

// CommitTree creates a commit from a tree and message, optionally with parents.
func (repo *TestRepo) CommitTree(tb testing.TB, tree oid.ObjectID, message string, parents ...oid.ObjectID) oid.ObjectID {
	tb.Helper()
	args := []string{"commit-tree", tree.String()}
	for _, p := range parents {
		args = append(args, "-p", p.String())
	}
	args = append(args, "-m", message)
	hex := repo.Run(tb, args...)
	id, err := oid.ParseHex(repo.algo, hex)
	if err != nil {
		tb.Fatalf("parse commit-tree output %q: %v", hex, err)
	}
	return id
}
