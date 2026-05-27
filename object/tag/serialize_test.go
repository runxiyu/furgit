package tag_test

import (
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	objectid "lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/tag"
)

func TestTagSerialize(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		_, _, commitID := testRepo.MakeCommit(t, "subject\n\nbody")
		tagID := testRepo.TagAnnotated(t, "v1", commitID, "tag message")

		rawBody := testRepo.CatFile(t, "tag", tagID)

		parsed, err := tag.Parse(rawBody, algo)
		if err != nil {
			t.Fatalf("ParseTag: %v", err)
		}

		rawObj, err := parsed.BytesWithHeader()
		if err != nil {
			t.Fatalf("BytesWithHeader: %v", err)
		}

		gotID := algo.Sum(rawObj)
		if gotID != tagID {
			t.Fatalf("tag id mismatch: got %s want %s", gotID, tagID)
		}
	})
}
