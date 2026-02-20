package testgit

import (
	"fmt"
	"testing"

	"codeberg.org/lindenii/furgit/objectid"
)

// TagAnnotated creates an annotated tag object and returns the resulting tag object ID.
func (repo *TestRepo) TagAnnotated(tb testing.TB, name string, target objectid.ObjectID, message string) objectid.ObjectID {
	tb.Helper()
	repo.Run(tb, "tag", "-a", name, target.String(), "-m", message)
	return repo.RevParse(tb, fmt.Sprintf("refs/tags/%s", name))
}
