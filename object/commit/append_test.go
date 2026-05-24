package commit_test

import (
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/object/commit"
	"codeberg.org/lindenii/furgit/object/id"
)

func TestCommitSerialize(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo id.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		_, _, commitID := testRepo.MakeCommit(t, "subject\n\nbody")

		rawBody := testRepo.CatFile(t, "commit", commitID)

		parsed, err := commit.Parse(rawBody, algo)
		if err != nil {
			t.Fatalf("ParseCommit: %v", err)
		}

		rawObj, err := parsed.AppendWithHeader([]byte(nil))
		if err != nil {
			t.Fatalf("BytesWithHeader: %v", err)
		}

		gotID := algo.Sum(rawObj)
		if gotID != commitID {
			t.Fatalf("commit id mismatch: got %s want %s", gotID, commitID)
		}
	})
}
