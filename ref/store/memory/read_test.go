package memory_test

import (
	"errors"
	"testing"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/ref"
	"lindenii.org/go/furgit/ref/store"
	"lindenii.org/go/furgit/ref/store/memory"
)

func TestResolveSymbolic(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			m := memory.New(objectFormat)
			mainID := objectFormat.Sum([]byte("main"))

			seed(t, m, func(tx store.Transaction) {
				err := tx.Create("refs/heads/main", mainID)
				if err != nil {
					t.Fatalf("Create(main): %v", err)
				}

				err = tx.CreateSymbolic("HEAD", "refs/heads/main")
				if err != nil {
					t.Fatalf("CreateSymbolic(HEAD): %v", err)
				}
			})

			head, err := m.Resolve("HEAD")
			if err != nil {
				t.Fatalf("Resolve(HEAD): %v", err)
			}

			symbolic, ok := head.(ref.Symbolic)
			if !ok {
				t.Fatalf("Resolve(HEAD) = %T, want ref.Symbolic", head)
			}

			if symbolic.Target != "refs/heads/main" {
				t.Fatalf("HEAD target = %q, want refs/heads/main", symbolic.Target)
			}

			direct, err := m.ResolveToDirect("HEAD")
			if err != nil {
				t.Fatalf("ResolveToDirect(HEAD): %v", err)
			}

			if direct.ID != mainID {
				t.Fatalf("ResolveToDirect(HEAD) ID = %v, want %v", direct.ID, mainID)
			}
		})
	}
}

func TestResolveMissing(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			m := memory.New(objectFormat)

			_, err := m.Resolve("refs/heads/absent")
			if err == nil {
				t.Fatalf("Resolve(absent) succeeded, want ErrReferenceNotFound")
			}

			if !errors.Is(err, store.ErrReferenceNotFound) {
				t.Fatalf("Resolve(absent) err = %v, want ErrReferenceNotFound", err)
			}
		})
	}
}
