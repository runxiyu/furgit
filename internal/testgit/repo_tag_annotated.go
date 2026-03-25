package testgit

import (
	"testing"

	objectid "codeberg.org/lindenii/furgit/object/id"
)

// TagAnnotated creates an annotated tag object and returns the resulting tag object ID.
func (testRepo *TestRepo) TagAnnotated(tb testing.TB, name string, target objectid.ObjectID, message string) objectid.ObjectID {
	tb.Helper()
	testRepo.Run(tb, "tag", "-a", name, target.String(), "-m", message)

	return testRepo.RevParse(tb, "refs/tags/"+name)
}
