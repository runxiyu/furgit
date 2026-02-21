package object_test

import (
	"bytes"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
)

func TestTagParseFromGit(t *testing.T) {
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		_, _, commitID := testRepo.MakeCommit(t, "subject\n\nbody")
		tagID := testRepo.TagAnnotated(t, "v1", commitID, "tag message")

		rawBody := testRepo.CatFile(t, "tag", tagID)
		tag, err := object.ParseTag(rawBody, algo)
		if err != nil {
			t.Fatalf("ParseTag: %v", err)
		}
		if tag.Target != commitID {
			t.Fatalf("tag target mismatch: got %s want %s", tag.Target, commitID)
		}
		if tag.TargetType != objecttype.TypeCommit {
			t.Fatalf("tag target type = %v, want %v", tag.TargetType, objecttype.TypeCommit)
		}
		if !bytes.Equal(tag.Name, []byte("v1")) {
			t.Fatalf("tag name = %q, want %q", tag.Name, "v1")
		}
		if tag.Tagger == nil {
			t.Fatalf("expected tagger")
		}
		if !bytes.Contains(tag.Message, []byte("tag message")) {
			t.Fatalf("tag message mismatch: %q", tag.Message)
		}
	})
}
