package testgit

import (
	"testing"

	objectid "codeberg.org/lindenii/furgit/object/id"
)

// CatFile returns raw output from git cat-file.
func (testRepo *TestRepo) CatFile(tb testing.TB, mode string, id objectid.ObjectID) []byte {
	tb.Helper()

	return testRepo.RunBytes(tb, "cat-file", mode, id.String())
}
