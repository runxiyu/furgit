package memory_test

import (
	"errors"
	"testing"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/ref/store"
	"lindenii.org/go/furgit/ref/store/memory"
)

func TestTransactionRejectLeavesStoreUnchanged(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			m := memory.New(objectFormat)
			mainID := objectFormat.Sum([]byte("main"))
			devID := objectFormat.Sum([]byte("dev"))
			nextID := objectFormat.Sum([]byte("next"))
			wrongOld := objectFormat.Sum([]byte("wrong"))

			seed(t, m, func(tx store.Transaction) {
				err := tx.Create("refs/heads/main", mainID)
				if err != nil {
					t.Fatalf("Create(main): %v", err)
				}

				err = tx.Create("refs/heads/dev", devID)
				if err != nil {
					t.Fatalf("Create(dev): %v", err)
				}
			})

			tx, err := m.BeginTransaction()
			if err != nil {
				t.Fatalf("BeginTransaction: %v", err)
			}

			err = tx.Update("refs/heads/main", nextID, mainID)
			if err != nil {
				t.Fatalf("Update(main): %v", err)
			}

			err = tx.Update("refs/heads/dev", nextID, wrongOld)
			if err != nil {
				t.Fatalf("Update(dev): %v", err)
			}

			err = tx.Commit()
			if err == nil {
				t.Fatalf("Commit succeeded, want WrongOldIDError")
			}

			if _, ok := errors.AsType[*store.WrongOldIDError](err); !ok {
				t.Fatalf("Commit error = %T %v, want *store.WrongOldIDError", err, err)
			}

			if got := resolveDirect(t, m, "refs/heads/main").ID; got != mainID {
				t.Fatalf("main after rejected transaction = %v, want %v", got, mainID)
			}

			if got := resolveDirect(t, m, "refs/heads/dev").ID; got != devID {
				t.Fatalf("dev after rejected transaction = %v, want %v", got, devID)
			}
		})
	}
}

func TestTransactionRejectsForeignObjectFormat(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			m := memory.New(objectFormat)

			tx, err := m.BeginTransaction()
			if err != nil {
				t.Fatalf("BeginTransaction: %v", err)
			}

			err = tx.Create("refs/heads/main", id.ObjectID{})
			if err == nil {
				t.Fatalf("Create with unset ID succeeded, want ErrInvalidValue")
			}

			if !errors.Is(err, store.ErrInvalidValue) {
				t.Fatalf("Create error = %v, want ErrInvalidValue", err)
			}
		})
	}
}
