package testgit

import (
	"fmt"
	"testing"

	"codeberg.org/lindenii/furgit/objectid"
)

// TagAnnotated creates an annotated tag object and returns the resulting tag object ID.
func (testRepo *TestRepo) TagAnnotated(tb testing.TB, name string, target objectid.ObjectID, message string) objectid.ObjectID {
	tb.Helper()
	testRepo.Run(tb, "tag", "-a", name, target.String(), "-m", message)
	return testRepo.RevParse(tb, fmt.Sprintf("refs/tags/%s", name))
}
