package commit_test

import (
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/object/commit"
	objectid "lindenii.org/go/furgit/object/id"
)

func TestCommitSerialize(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		_, _, commitID := testRepo.MakeCommit(t, "subject\n\nbody")

		rawBody := testRepo.CatFile(t, "commit", commitID)

		parsed, err := commit.Parse(rawBody, algo)
		if err != nil {
			t.Fatalf("ParseCommit: %v", err)
		}

		rawObj, err := parsed.BytesWithHeader()
		if err != nil {
			t.Fatalf("BytesWithHeader: %v", err)
		}

		gotID := algo.Sum(rawObj)
		if gotID != commitID {
			t.Fatalf("commit id mismatch: got %s want %s", gotID, commitID)
		}
	})
}
