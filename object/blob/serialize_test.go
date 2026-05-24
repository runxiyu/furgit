package blob_test

import (
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/object/blob"
	objectid "codeberg.org/lindenii/furgit/object/id"
)

func TestBlobSerialize(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		body := []byte("hello\nblob\n")
		wantID := testRepo.HashObject(t, "blob", body)

		obj := &blob.Blob{Data: body}

		rawObj, err := obj.AppendWithHeader([]byte(nil))
		if err != nil {
			t.Fatalf("BytesWithHeader: %v", err)
		}

		gotID := algo.Sum(rawObj)
		if gotID != wantID {
			t.Fatalf("object id mismatch: got %s want %s", gotID, wantID)
		}
	})
}
