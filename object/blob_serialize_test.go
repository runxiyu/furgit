package object_test

import (
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/objectid"
)

func TestBlobSerialize(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		body := []byte("hello\nblob\n")
		wantID := testRepo.HashObject(t, "blob", body)

		blob := &object.Blob{Data: body}
		rawObj, err := blob.SerializeWithHeader()
		if err != nil {
			t.Fatalf("SerializeWithHeader: %v", err)
		}
		gotID := algo.Sum(rawObj)
		if gotID != wantID {
			t.Fatalf("object id mismatch: got %s want %s", gotID, wantID)
		}
	})
}
