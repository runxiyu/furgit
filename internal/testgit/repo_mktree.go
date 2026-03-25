package testgit

import (
	"testing"

	objectid "codeberg.org/lindenii/furgit/object/id"
)

// Mktree creates a tree from textual mktree input and returns its ID.
func (testRepo *TestRepo) Mktree(tb testing.TB, input string) objectid.ObjectID {
	tb.Helper()
	hex := testRepo.RunInput(tb, []byte(input), "mktree")

	id, err := objectid.ParseHex(testRepo.algo, hex)
	if err != nil {
		tb.Fatalf("parse mktree output %q: %v", hex, err)
	}

	return id
}
