package object_test

import (
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/objectid"
)

func TestBlobSerialize(t *testing.T) {
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		testRepo := testgit.NewBareRepo(t, algo)
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
