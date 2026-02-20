package object_test

import (
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/oid"
)

func TestCommitSerialize(t *testing.T) {
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo oid.Algorithm) {
		repo := testgit.NewBareRepo(t, algo)
		_, _, commitID := repo.MakeCommit(t, "subject\n\nbody")

		rawBody := repo.CatFile(t, "commit", commitID)
		commit, err := object.ParseCommit(rawBody, algo)
		if err != nil {
			t.Fatalf("ParseCommit: %v", err)
		}

		rawObj, err := commit.Serialize()
		if err != nil {
			t.Fatalf("Serialize: %v", err)
		}
		gotID := algo.Sum(rawObj)
		if gotID != commitID {
			t.Fatalf("commit id mismatch: got %s want %s", gotID, commitID)
		}
	})
}
