package tree_test

import (
	"bytes"
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/object/id"
)

func TestFind(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo, err := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: objectFormat})
			if err != nil {
				t.Fatalf("NewRepo: %v", err)
			}

			entries := mixedEntries(t, repo)
			tr := buildTree(t, entries)

			for _, want := range entries {
				got, ok := tr.Find(want.Name)
				if !ok {
					t.Fatalf("Find(%q) not found", want.Name)
				}

				if got.Mode != want.Mode || !bytes.Equal(got.Name, want.Name) || got.ID != want.ID {
					t.Fatalf("Find(%q) = %+v, want %+v", want.Name, got, want)
				}
			}

			if _, ok := tr.Find([]byte("does-not-exist")); ok {
				t.Fatalf("Find(does-not-exist) = true, want false")
			}
		})
	}
}
