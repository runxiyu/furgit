package tag_test

import (
	"bytes"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/tag"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

func TestTagParseFromGit(t *testing.T) {
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

		if parsed.TargetID != commitID {
			t.Fatalf("tag target mismatch: got %s want %s", parsed.TargetID, commitID)
		}

		if parsed.TargetType != objecttype.TypeCommit {
			t.Fatalf("tag target type = %v, want %v", parsed.TargetType, objecttype.TypeCommit)
		}

		if !bytes.Equal(parsed.Name, []byte("v1")) {
			t.Fatalf("tag name = %q, want %q", parsed.Name, "v1")
		}

		if parsed.Tagger == nil {
			t.Fatalf("expected tagger")
		}

		if !bytes.Contains(parsed.Message, []byte("tag message")) {
			t.Fatalf("tag message mismatch: %q", parsed.Message)
		}
	})
}
