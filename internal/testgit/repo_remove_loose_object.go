package testgit

import (
	"fmt"
	"testing"

	objectid "lindenii.org/go/furgit/object/id"
)

// RemoveLooseObject removes one loose object file from the repository.
func (testRepo *TestRepo) RemoveLooseObject(tb testing.TB, id objectid.ObjectID) {
	tb.Helper()

	root := testRepo.OpenObjectsRoot(tb)
	hex := id.String()
	path := fmt.Sprintf("%s/%s", hex[:2], hex[2:])

	err := root.Remove(path)
	if err != nil {
		tb.Fatalf("remove loose object %s: %v", id, err)
	}
}
