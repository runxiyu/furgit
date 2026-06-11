package loose_test

import (
	"bytes"
	"errors"
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store"
	"lindenii.org/go/furgit/object/typ"
)

func TestQuarantinePromote(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo, err := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: objectFormat})
			if err != nil {
				t.Fatalf("NewRepo: %v", err)
			}

			looseStore := openLooseStore(t, repo)

			quarantine, err := looseStore.BeginObjectQuarantine(store.ObjectQuarantineOptions{})
			if err != nil {
				t.Fatalf("BeginObjectQuarantine: %v", err)
			}

			content := []byte("quarantined object\n")

			objectID, err := quarantine.WriteBytesContent(typ.Blob, content)
			if err != nil {
				t.Fatalf("quarantine.WriteBytesContent: %v", err)
			}

			ty, got, err := quarantine.ReadBytesContent(objectID)
			if err != nil {
				t.Fatalf("quarantine.ReadBytesContent: %v", err)
			}

			if ty != typ.Blob {
				t.Fatalf("quarantine type = %v, want %v", ty, typ.Blob)
			}

			if !bytes.Equal(got, content) {
				t.Fatalf("quarantine body mismatch")
			}

			_, _, err = looseStore.ReadBytesContent(objectID)
			if !errors.Is(err, store.ErrObjectNotFound) {
				t.Fatalf("parent saw quarantined object before promote: %v", err)
			}

			err = quarantine.Promote()
			if err != nil {
				t.Fatalf("Promote: %v", err)
			}

			ty, got, err = looseStore.ReadBytesContent(objectID)
			if err != nil {
				t.Fatalf("parent ReadBytesContent after promote: %v", err)
			}

			if ty != typ.Blob {
				t.Fatalf("parent type = %v, want %v", ty, typ.Blob)
			}

			if !bytes.Equal(got, content) {
				t.Fatalf("parent body mismatch")
			}
		})
	}
}

func TestQuarantineDiscard(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo, err := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: objectFormat})
			if err != nil {
				t.Fatalf("NewRepo: %v", err)
			}

			looseStore := openLooseStore(t, repo)

			quarantine, err := looseStore.BeginObjectQuarantine(store.ObjectQuarantineOptions{})
			if err != nil {
				t.Fatalf("BeginObjectQuarantine: %v", err)
			}

			content := []byte("discarded object\n")

			objectID, err := quarantine.WriteBytesContent(typ.Blob, content)
			if err != nil {
				t.Fatalf("quarantine.WriteBytesContent: %v", err)
			}

			err = quarantine.Discard()
			if err != nil {
				t.Fatalf("Discard: %v", err)
			}

			_, _, err = looseStore.ReadBytesContent(objectID)
			if !errors.Is(err, store.ErrObjectNotFound) {
				t.Fatalf("parent saw discarded object: %v", err)
			}
		})
	}
}
