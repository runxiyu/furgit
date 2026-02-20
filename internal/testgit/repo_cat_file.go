package testgit

import (
	"testing"

	"codeberg.org/lindenii/furgit/oid"
)

// CatFile returns raw output from git cat-file.
func (repo *TestRepo) CatFile(tb testing.TB, mode string, id oid.ObjectID) []byte {
	tb.Helper()
	return repo.RunBytes(tb, "cat-file", mode, id.String())
}
