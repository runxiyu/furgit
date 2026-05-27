package testgit

import (
	"fmt"
	"testing"

	objectid "lindenii.org/go/furgit/object/id"
)

// MakeSingleFileTree writes one blob and one tree entry for it and returns (blobID, treeID).
func (testRepo *TestRepo) MakeSingleFileTree(tb testing.TB, fileName string, fileContent []byte) (objectid.ObjectID, objectid.ObjectID) {
	tb.Helper()
	blobID := testRepo.HashObject(tb, "blob", fileContent)
	treeInput := fmt.Sprintf("100644 blob %s\t%s\n", blobID.String(), fileName)
	treeID := testRepo.Mktree(tb, treeInput)

	return blobID, treeID
}
