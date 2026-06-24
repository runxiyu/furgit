package memory_test

import (
	"errors"
	"testing"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/ref"
	"lindenii.org/go/furgit/ref/store"
	"lindenii.org/go/furgit/ref/store/memory"
)

func TestBatchRejectsDuplicateResolvedTargetAndAppliesRemainder(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			m := memory.New(objectFormat)
			mainID := objectFormat.Sum([]byte("main"))
			devID := objectFormat.Sum([]byte("dev"))
			nextMainID := objectFormat.Sum([]byte("next-main"))
			nextDevID := objectFormat.Sum([]byte("next-dev"))
			aliasID := objectFormat.Sum([]byte("alias"))

			seed(t, m, func(tx store.Transaction) {
				err := tx.Create("refs/heads/main", mainID)
				if err != nil {
					t.Fatalf("Create(main): %v", err)
				}

				err = tx.Create("refs/heads/dev", devID)
				if err != nil {
					t.Fatalf("Create(dev): %v", err)
				}

				err = tx.CreateSymbolic("refs/heads/alias", "refs/heads/main")
				if err != nil {
					t.Fatalf("CreateSymbolic(alias): %v", err)
				}
			})

			batch, err := m.BeginBatch()
			if err != nil {
				t.Fatalf("BeginBatch: %v", err)
			}

			err = batch.Update("refs/heads/main", nextMainID, mainID)
			if err != nil {
				t.Fatalf("Update(main): %v", err)
			}

			// Updates the symbolic alias in deref mode,
			// which resolves to refs/heads/main
			// and therefore duplicates the first operation.
			err = batch.Update("refs/heads/alias", aliasID, mainID)
			if err != nil {
				t.Fatalf("Update(alias): %v", err)
			}

			err = batch.Update("refs/heads/dev", nextDevID, devID)
			if err != nil {
				t.Fatalf("Update(dev): %v", err)
			}

			results, err := batch.Apply()
			if err != nil {
				t.Fatalf("Apply: %v", err)
			}

			if len(results) != 3 {
				t.Fatalf("len(results) = %d, want 3", len(results))
			}

			if results[0].Status != store.BatchStatusApplied {
				t.Fatalf("results[0].Status = %v, want applied", results[0].Status)
			}

			if results[1].Status != store.BatchStatusRejected {
				t.Fatalf("results[1].Status = %v, want rejected", results[1].Status)
			}

			if !errors.Is(results[1].Error, store.ErrDuplicateUpdate) {
				t.Fatalf("results[1].Error = %v, want ErrDuplicateUpdate", results[1].Error)
			}

			if results[2].Status != store.BatchStatusApplied {
				t.Fatalf("results[2].Status = %v, want applied", results[2].Status)
			}

			if got := resolveDirect(t, m, "refs/heads/main").ID; got != nextMainID {
				t.Fatalf("main after batch = %v, want %v", got, nextMainID)
			}

			if got := resolveDirect(t, m, "refs/heads/dev").ID; got != nextDevID {
				t.Fatalf("dev after batch = %v, want %v", got, nextDevID)
			}

			resolved, err := m.Resolve("refs/heads/alias")
			if err != nil {
				t.Fatalf("Resolve(alias): %v", err)
			}

			symbolic, ok := resolved.(ref.Symbolic)
			if !ok {
				t.Fatalf("Resolve(alias) = %T, want ref.Symbolic", resolved)
			}

			if symbolic.Target != "refs/heads/main" {
				t.Fatalf("alias target = %q, want refs/heads/main", symbolic.Target)
			}
		})
	}
}
