package testgit

import (
	"testing"

	objectid "codeberg.org/lindenii/furgit/object/id"
)

// CommitTree creates a commit from a tree and message, optionally with parents.
func (testRepo *TestRepo) CommitTree(tb testing.TB, tree objectid.ObjectID, message string, parents ...objectid.ObjectID) objectid.ObjectID {
	tb.Helper()

	args := make([]string, 0, 2+2*len(parents)+2)

	args = append(args, "commit-tree", tree.String())
	for _, p := range parents {
		args = append(args, "-p", p.String())
	}

	args = append(args, "-m", message)
	hex := testRepo.Run(tb, args...)

	id, err := objectid.ParseHex(testRepo.algo, hex)
	if err != nil {
		tb.Fatalf("parse commit-tree output %q: %v", hex, err)
	}

	return id
}
