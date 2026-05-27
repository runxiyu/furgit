package testgit

import (
	"fmt"
	"strings"
	"testing"

	objectid "lindenii.org/go/furgit/object/id"
)

const (
	manyObjectsMainCommits = 640
	manyObjectsDevCommits  = 220
)

// MakeManyObjectsHistory creates a large commit graph.
func (testRepo *TestRepo) MakeManyObjectsHistory(tb testing.TB) {
	tb.Helper()

	var (
		mainTip objectid.ObjectID
		devTip  objectid.ObjectID
		hasMain bool
		hasDev  bool
	)

	for i := range manyObjectsMainCommits {
		tree := testRepo.makeManyObjectsTree(tb, "main", i, 3)

		var commit objectid.ObjectID
		if hasMain {
			commit = testRepo.CommitTree(tb, tree, fmt.Sprintf("main-%04d", i), mainTip)
		} else {
			commit = testRepo.CommitTree(tb, tree, fmt.Sprintf("main-%04d", i))
			hasMain = true
		}

		mainTip = commit
		if i%64 == 0 {
			testRepo.TagAnnotated(tb, fmt.Sprintf("main-v%04d", i), mainTip, fmt.Sprintf("tag-main-%04d", i))
		}
	}

	devTip = mainTip
	hasDev = true

	for i := range manyObjectsDevCommits {
		tree := testRepo.makeManyObjectsTree(tb, "dev", i, 4)
		commit := testRepo.CommitTree(tb, tree, fmt.Sprintf("dev-%04d", i), devTip)
		devTip = commit

		if i > 0 && i%55 == 0 {
			mergeTree := testRepo.makeManyObjectsTree(tb, "merge", i, 2)

			mainTip = testRepo.CommitTree(tb, mergeTree, fmt.Sprintf("merge-%04d", i), mainTip, devTip)
			if i%110 == 0 {
				testRepo.TagAnnotated(tb, fmt.Sprintf("merge-v%04d", i), mainTip, fmt.Sprintf("tag-merge-%04d", i))
			}
		}
	}

	if hasMain {
		testRepo.UpdateRef(tb, "refs/heads/main", mainTip)
	}

	if hasDev {
		testRepo.UpdateRef(tb, "refs/heads/dev", devTip)
	}
}

// makeManyObjectsTree builds one synthetic tree with fanout blobs.
func (testRepo *TestRepo) makeManyObjectsTree(tb testing.TB, prefix string, i int, files int) objectid.ObjectID {
	tb.Helper()

	lines := make([]string, 0, files)
	for j := range files {
		body := fmt.Appendf(nil, "%s-%04d-%02d\n%s\n", prefix, i, j, strings.Repeat("x", 160+(i+j)%96))
		blobID := testRepo.HashObject(tb, "blob", body)
		lines = append(lines, fmt.Sprintf("100644 blob %s\t%s_%04d_%02d.txt\n", blobID.String(), prefix, i, j))
	}

	return testRepo.Mktree(tb, strings.Join(lines, ""))
}
