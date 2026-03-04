package testgit

import (
	"testing"

	"codeberg.org/lindenii/furgit/objectid"
)

// CatFile returns raw output from git cat-file.
func (testRepo *TestRepo) CatFile(tb testing.TB, mode string, id objectid.ObjectID) []byte {
	tb.Helper()

	return testRepo.RunBytes(tb, "cat-file", mode, id.String())
}
