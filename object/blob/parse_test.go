package blob_test

import (
	"bytes"
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/object/blob"
	objectid "lindenii.org/go/furgit/object/id"
)

func TestBlobParseFromGit(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		body := []byte("hello\nblob\n")
		blobID := testRepo.HashObject(t, "blob", body)

		rawBody := testRepo.CatFile(t, "blob", blobID)

		parsed, err := blob.Parse(rawBody)
		if err != nil {
			t.Fatalf("ParseBlob: %v", err)
		}

		if !bytes.Equal(parsed.Data, body) {
			t.Fatalf("blob body mismatch")
		}
	})
}
