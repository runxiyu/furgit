package testgit

import (
	"testing"

	"codeberg.org/lindenii/furgit/objectid"
)

// Mktree creates a tree from textual mktree input and returns its ID.
func (repo *TestRepo) Mktree(tb testing.TB, input string) objectid.ObjectID {
	tb.Helper()
	hex := repo.RunInput(tb, []byte(input), "mktree")
	id, err := objectid.ParseHex(repo.algo, hex)
	if err != nil {
		tb.Fatalf("parse mktree output %q: %v", hex, err)
	}
	return id
}
