package object_test

import (
	"bytes"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/objecttype"
	"codeberg.org/lindenii/furgit/oid"
)

func TestTagParseFromGit(t *testing.T) {
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo oid.Algorithm) {
		repo := testgit.NewBareRepo(t, algo)
		_, _, commitID := repo.MakeCommit(t, "subject\n\nbody")
		tagID := repo.TagAnnotated(t, "v1", commitID, "tag message")

		rawBody := repo.CatFile(t, "tag", tagID)
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
