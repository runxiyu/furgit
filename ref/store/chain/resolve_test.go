package chain_test

import (
	"errors"
	"testing"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/ref/store"
	"lindenii.org/go/furgit/ref/store/chain"
	"lindenii.org/go/furgit/ref/store/memory"
)

func seed(t *testing.T, m *memory.Memory, fn func(tx store.Transaction)) {
	t.Helper()

	tx, err := m.BeginTransaction()
	if err != nil {
		t.Fatalf("BeginTransaction: %v", err)
	}

	fn(tx)

	err = tx.Commit()
	if err != nil {
		t.Fatalf("Commit: %v", err)
	}
}

func TestResolvePrecedence(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			frontID := objectFormat.Sum([]byte("front"))
			backID := objectFormat.Sum([]byte("back"))

			front := memory.New(objectFormat)
			back := memory.New(objectFormat)

			seed(t, front, func(tx store.Transaction) {
				err := tx.Create("refs/heads/main", frontID)
				if err != nil {
					t.Fatalf("front Create: %v", err)
				}
			})

			seed(t, back, func(tx store.Transaction) {
				err := tx.Create("refs/heads/main", backID)
				if err != nil {
					t.Fatalf("back Create: %v", err)
				}
			})

			c := chain.New(front, back)

			direct, err := c.ResolveToDirect("refs/heads/main")
			if err != nil {
				t.Fatalf("ResolveToDirect: %v", err)
			}

			if direct.ID != frontID {
				t.Fatalf("ID = %v, want front %v", direct.ID, frontID)
			}
		})
	}
}

func TestResolveFallThrough(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			backID := objectFormat.Sum([]byte("back"))

			front := memory.New(objectFormat)
			back := memory.New(objectFormat)

			seed(t, back, func(tx store.Transaction) {
				err := tx.Create("refs/heads/only-back", backID)
				if err != nil {
					t.Fatalf("back Create: %v", err)
				}
			})

			c := chain.New(front, back)

			direct, err := c.ResolveToDirect("refs/heads/only-back")
			if err != nil {
				t.Fatalf("ResolveToDirect: %v", err)
			}

			if direct.ID != backID {
				t.Fatalf("ID = %v, want back %v", direct.ID, backID)
			}
		})
	}
}

func TestResolveNotFound(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			c := chain.New(memory.New(objectFormat), memory.New(objectFormat))

			_, err := c.Resolve("refs/heads/missing")
			if !errors.Is(err, store.ErrReferenceNotFound) {
				t.Fatalf("Resolve err = %v, want ErrReferenceNotFound", err)
			}
		})
	}
}

func TestResolveToDirectAcrossBackends(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			mainID := objectFormat.Sum([]byte("main"))

			front := memory.New(objectFormat)
			back := memory.New(objectFormat)

			seed(t, front, func(tx store.Transaction) {
				err := tx.CreateSymbolic("HEAD", "refs/heads/main")
				if err != nil {
					t.Fatalf("front CreateSymbolic: %v", err)
				}
			})

			seed(t, back, func(tx store.Transaction) {
				err := tx.Create("refs/heads/main", mainID)
				if err != nil {
					t.Fatalf("back Create: %v", err)
				}
			})

			c := chain.New(front, back)

			direct, err := c.ResolveToDirect("HEAD")
			if err != nil {
				t.Fatalf("ResolveToDirect(HEAD): %v", err)
			}

			if direct.ID != mainID {
				t.Fatalf("ID = %v, want %v", direct.ID, mainID)
			}
		})
	}
}

func TestResolveToDirectCycleAcrossBackends(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			front := memory.New(objectFormat)
			back := memory.New(objectFormat)

			seed(t, front, func(tx store.Transaction) {
				err := tx.CreateSymbolic("refs/a", "refs/b")
				if err != nil {
					t.Fatalf("front CreateSymbolic: %v", err)
				}
			})

			seed(t, back, func(tx store.Transaction) {
				err := tx.CreateSymbolic("refs/b", "refs/a")
				if err != nil {
					t.Fatalf("back CreateSymbolic: %v", err)
				}
			})

			c := chain.New(front, back)

			_, err := c.ResolveToDirect("refs/a")
			if !errors.Is(err, store.ErrSymbolicCycle) {
				t.Fatalf("ResolveToDirect err = %v, want ErrSymbolicCycle", err)
			}
		})
	}
}
