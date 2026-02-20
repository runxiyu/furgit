package object_test

import (
	"bytes"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/objectid"
)

func TestBlobParseFromGit(t *testing.T) {
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		repo := testgit.NewBareRepo(t, algo)
		body := []byte("hello\nblob\n")
		blobID := repo.HashObject(t, "blob", body)

		rawBody := repo.CatFile(t, "blob", blobID)
		blob, err := object.ParseBlob(rawBody)
		if err != nil {
			t.Fatalf("ParseBlob: %v", err)
		}
		if !bytes.Equal(blob.Data, body) {
			t.Fatalf("blob body mismatch")
		}
	})
}
