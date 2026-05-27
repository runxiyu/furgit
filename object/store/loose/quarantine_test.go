package loose_test

import (
	"bytes"
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	objectid "lindenii.org/go/furgit/object/id"
	objectstore "lindenii.org/go/furgit/object/store"
	objecttype "lindenii.org/go/furgit/object/type"
)

func TestLooseQuarantinePromotePublishesWrittenObjects(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		store := openLooseStore(t, testRepo, algo)

		quarantiner, ok := any(store).(objectstore.ObjectQuarantiner)
		if !ok {
			t.Fatal("loose store does not implement ObjectQuarantiner")
		}

		quarantine, err := quarantiner.BeginObjectQuarantine(objectstore.ObjectQuarantineOptions{})
		if err != nil {
			t.Fatalf("BeginObjectQuarantine: %v", err)
		}

		content := []byte("quarantined loose object\n")

		id, err := quarantine.WriteBytesContent(objecttype.TypeBlob, content)
		if err != nil {
			t.Fatalf("quarantine.WriteBytesContent: %v", err)
		}

		ty, got, err := quarantine.ReadBytesContent(id)
		if err != nil {
			t.Fatalf("quarantine.ReadBytesContent: %v", err)
		}

		if ty != objecttype.TypeBlob {
			t.Fatalf("quarantine.ReadBytesContent type = %v, want %v", ty, objecttype.TypeBlob)
		}

		if !bytes.Equal(got, content) {
			t.Fatal("quarantine.ReadBytesContent mismatch")
		}

		_, _, err = store.ReadBytesContent(id)
		if err == nil {
			t.Fatal("store.ReadBytesContent unexpectedly saw quarantined object before promote")
		}

		err = quarantine.Promote()
		if err != nil {
			t.Fatalf("quarantine.Promote: %v", err)
		}

		err = store.Refresh()
		if err != nil {
			t.Fatalf("store.Refresh: %v", err)
		}

		ty, got, err = store.ReadBytesContent(id)
		if err != nil {
			t.Fatalf("store.ReadBytesContent after promote: %v", err)
		}

		if ty != objecttype.TypeBlob {
			t.Fatalf("store.ReadBytesContent type = %v, want %v", ty, objecttype.TypeBlob)
		}

		if !bytes.Equal(got, content) {
			t.Fatal("store.ReadBytesContent mismatch")
		}
	})
}

func TestLooseQuarantineDiscardDropsWrittenObjects(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		store := openLooseStore(t, testRepo, algo)

		quarantiner, ok := any(store).(objectstore.ObjectQuarantiner)
		if !ok {
			t.Fatal("expected objectstore.ObjectQuarantiner")
		}

		quarantine, err := quarantiner.BeginObjectQuarantine(objectstore.ObjectQuarantineOptions{})
		if err != nil {
			t.Fatalf("BeginObjectQuarantine: %v", err)
		}

		content := []byte("discarded loose object\n")

		id, err := quarantine.WriteBytesContent(objecttype.TypeBlob, content)
		if err != nil {
			t.Fatalf("quarantine.WriteBytesContent: %v", err)
		}

		err = quarantine.Discard()
		if err != nil {
			t.Fatalf("quarantine.Discard: %v", err)
		}

		err = store.Refresh()
		if err != nil {
			t.Fatalf("store.Refresh: %v", err)
		}

		_, _, err = store.ReadBytesContent(id)
		if err == nil {
			t.Fatal("store.ReadBytesContent unexpectedly saw discarded object")
		}
	})
}
