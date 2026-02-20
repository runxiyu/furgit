package testgit

import (
	"testing"

	"codeberg.org/lindenii/furgit/objectid"
)

// MakeCommit creates a commit over a single-file tree and returns (blobID, treeID, commitID).
func (repo *TestRepo) MakeCommit(tb testing.TB, message string) (objectid.ObjectID, objectid.ObjectID, objectid.ObjectID) {
	tb.Helper()
	blobID, treeID := repo.MakeSingleFileTree(tb, "file.txt", []byte("commit-body\n"))
	commitID := repo.CommitTree(tb, treeID, message)
	return blobID, treeID, commitID
}
