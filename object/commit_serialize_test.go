package object_test

import (
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/objectid"
)

func TestCommitSerialize(t *testing.T) {
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		testRepo := testgit.NewBareRepo(t, algo)
		_, _, commitID := testRepo.MakeCommit(t, "subject\n\nbody")

		rawBody := testRepo.CatFile(t, "commit", commitID)
		commit, err := object.ParseCommit(rawBody, algo)
		if err != nil {
			t.Fatalf("ParseCommit: %v", err)
		}

		rawObj, err := commit.SerializeWithHeader()
		if err != nil {
			t.Fatalf("SerializeWithHeader: %v", err)
		}
		gotID := algo.Sum(rawObj)
		if gotID != commitID {
			t.Fatalf("commit id mismatch: got %s want %s", gotID, commitID)
		}
	})
}
