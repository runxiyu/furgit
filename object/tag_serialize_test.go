package object_test

import (
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/objectid"
)

func TestTagSerialize(t *testing.T) {
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		testRepo := testgit.NewBareRepo(t, algo)
		_, _, commitID := testRepo.MakeCommit(t, "subject\n\nbody")
		tagID := testRepo.TagAnnotated(t, "v1", commitID, "tag message")

		rawBody := testRepo.CatFile(t, "tag", tagID)
		tag, err := object.ParseTag(rawBody, algo)
		if err != nil {
			t.Fatalf("ParseTag: %v", err)
		}

		rawObj, err := tag.SerializeWithHeader()
		if err != nil {
			t.Fatalf("SerializeWithHeader: %v", err)
		}
		gotID := algo.Sum(rawObj)
		if gotID != tagID {
			t.Fatalf("tag id mismatch: got %s want %s", gotID, tagID)
		}
	})
}
