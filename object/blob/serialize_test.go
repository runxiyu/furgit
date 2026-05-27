package blob_test

import (
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/object/blob"
	objectid "lindenii.org/go/furgit/object/id"
)

func TestBlobSerialize(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		body := []byte("hello\nblob\n")
		wantID := testRepo.HashObject(t, "blob", body)

		obj := &blob.Blob{Data: body}

		rawObj, err := obj.BytesWithHeader()
		if err != nil {
			t.Fatalf("BytesWithHeader: %v", err)
		}

		gotID := algo.Sum(rawObj)
		if gotID != wantID {
			t.Fatalf("object id mismatch: got %s want %s", gotID, wantID)
		}
	})
}
