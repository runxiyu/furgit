package testgit

import (
	"testing"

	"codeberg.org/lindenii/furgit/oid"
)

// Mktree creates a tree from textual mktree input and returns its ID.
func (repo *TestRepo) Mktree(tb testing.TB, input string) oid.ObjectID {
	tb.Helper()
	hex := repo.RunInput(tb, []byte(input), "mktree")
	id, err := oid.ParseHex(repo.algo, hex)
	if err != nil {
		tb.Fatalf("parse mktree output %q: %v", hex, err)
	}
	return id
}
