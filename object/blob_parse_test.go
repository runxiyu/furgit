package object_test

import (
	"bytes"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/objectid"
)

func TestBlobParseFromGit(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		body := []byte("hello\nblob\n")
		blobID := testRepo.HashObject(t, "blob", body)

		rawBody := testRepo.CatFile(t, "blob", blobID)
		blob, err := object.ParseBlob(rawBody)
		if err != nil {
			t.Fatalf("ParseBlob: %v", err)
		}
		if !bytes.Equal(blob.Data, body) {
			t.Fatalf("blob body mismatch")
		}
	})
}
