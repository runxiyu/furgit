package object_test

import (
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/oid"
)

func TestTagSerialize(t *testing.T) {
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo oid.Algorithm) {
		repo := testgit.NewBareRepo(t, algo)
		_, _, commitID := repo.MakeCommit(t, "subject\n\nbody")
		tagID := repo.TagAnnotated(t, "v1", commitID, "tag message")

		rawBody := repo.CatFile(t, "tag", tagID)
		tag, err := object.ParseTag(rawBody, algo)
		if err != nil {
			t.Fatalf("ParseTag: %v", err)
		}

		rawObj, err := tag.Serialize()
		if err != nil {
			t.Fatalf("Serialize: %v", err)
		}
		gotID := algo.Sum(rawObj)
		if gotID != tagID {
			t.Fatalf("tag id mismatch: got %s want %s", gotID, tagID)
		}
	})
}
