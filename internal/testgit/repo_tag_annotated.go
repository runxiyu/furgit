package testgit

import (
	"fmt"
	"testing"

	"codeberg.org/lindenii/furgit/oid"
)

// TagAnnotated creates an annotated tag object and returns the resulting tag object ID.
func (repo *TestRepo) TagAnnotated(tb testing.TB, name string, target oid.ObjectID, message string) oid.ObjectID {
	tb.Helper()
	repo.Run(tb, "tag", "-a", name, target.String(), "-m", message)
	return repo.RevParse(tb, fmt.Sprintf("refs/tags/%s", name))
}
