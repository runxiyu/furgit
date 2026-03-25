package testgit

import (
	"testing"

	objectid "codeberg.org/lindenii/furgit/object/id"
)

// MakeCommit creates a commit over a single-file tree and returns (blobID, treeID, commitID).
func (testRepo *TestRepo) MakeCommit(tb testing.TB, message string) (objectid.ObjectID, objectid.ObjectID, objectid.ObjectID) {
	tb.Helper()
	blobID, treeID := testRepo.MakeSingleFileTree(tb, "file.txt", []byte("commit-body\n"))
	commitID := testRepo.CommitTree(tb, treeID, message)

	return blobID, treeID, commitID
}
