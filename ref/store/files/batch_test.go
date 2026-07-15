package files_test

import (
	"errors"
	"testing"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/ref"
	"lindenii.org/go/furgit/ref/store"
)

func TestBatchRejectsDuplicateResolvedTargetAndAppliesRemainder(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo := newFilesRepo(t, objectFormat)
			mainID := makeCommit(t, repo, "main")
			devID := makeCommit(t, repo, "dev")
			nextMainID := makeCommit(t, repo, "next-main")
			nextDevID := makeCommit(t, repo, "next-dev")
			aliasID := makeCommit(t, repo, "alias")

			err := repo.UpdateRef(t, "refs/heads/main", mainID)
			if err != nil {
				t.Fatalf("UpdateRef(main): %v", err)
			}

			err = repo.UpdateRef(t, "refs/heads/dev", devID)
			if err != nil {
				t.Fatalf("UpdateRef(dev): %v", err)
			}

			err = repo.SymbolicRefUpdate(t, "refs/heads/alias", "refs/heads/main")
			if err != nil {
				t.Fatalf("SymbolicRefUpdate(alias): %v", err)
			}

			filesStore := openStore(t, repo, objectFormat)

			batch, err := filesStore.BeginBatch()
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

			if got := gitResolve(t, repo, "refs/heads/main"); got != nextMainID {
				t.Fatalf("main after batch = %v, want %v", got, nextMainID)
			}

			if got := gitResolve(t, repo, "refs/heads/dev"); got != nextDevID {
				t.Fatalf("dev after batch = %v, want %v", got, nextDevID)
			}

			resolved, err := filesStore.Resolve("refs/heads/alias")
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

func TestBatchRejectsInvalidOpsAndAppliesRemainder(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo := newFilesRepo(t, objectFormat)
			mainID := makeCommit(t, repo, "main")
			devID := makeCommit(t, repo, "dev")
			newID := makeCommit(t, repo, "new")
			wrongID := makeCommit(t, repo, "wrong")

			err := repo.UpdateRef(t, "refs/heads/main", mainID)
			if err != nil {
				t.Fatalf("UpdateRef: %v", err)
			}

			err = repo.UpdateRef(t, "refs/heads/dev", devID)
			if err != nil {
				t.Fatalf("UpdateRef(dev): %v", err)
			}

			filesStore := openStore(t, repo, objectFormat)

			batch, err := filesStore.BeginBatch()
			if err != nil {
				t.Fatalf("BeginBatch: %v", err)
			}

			err = batch.Create("refs/heads/fresh", newID)
			if err != nil {
				t.Fatalf("Create(fresh): %v", err)
			}

			// refs/heads/main already exists.
			err = batch.Create("refs/heads/main", newID)
			if err != nil {
				t.Fatalf("Create(main): %v", err)
			}

			// wrong expected old value.
			err = batch.Update("refs/heads/dev", newID, wrongID)
			if err != nil {
				t.Fatalf("Update(dev): %v", err)
			}

			// target does not exist.
			err = batch.Delete("refs/heads/missing", wrongID)
			if err != nil {
				t.Fatalf("Delete(missing): %v", err)
			}

			results, err := batch.Apply()
			if err != nil {
				t.Fatalf("Apply: %v", err)
			}

			expected := []store.BatchStatus{
				store.BatchStatusApplied,
				store.BatchStatusRejected,
				store.BatchStatusRejected,
				store.BatchStatusRejected,
			}

			for i, status := range expected {
				if results[i].Status != status {
					t.Fatalf("results[%d] = %v (%v), want %v", i, results[i].Status, results[i].Error, status)
				}
			}

			if !errors.Is(results[1].Error, store.ErrCreateExists) {
				t.Fatalf("results[1].Error = %v, want ErrCreateExists", results[1].Error)
			}

			if _, ok := errors.AsType[*store.WrongOldIDError](results[2].Error); !ok {
				t.Fatalf("results[2].Error = %v, want WrongOldIDError", results[2].Error)
			}

			if !errors.Is(results[3].Error, store.ErrReferenceNotFound) {
				t.Fatalf("results[3].Error = %v, want ErrReferenceNotFound", results[3].Error)
			}

			if got := gitResolve(t, repo, "refs/heads/fresh"); got != newID {
				t.Fatalf("fresh after batch = %v, want %v", got, newID)
			}

			if got := gitResolve(t, repo, "refs/heads/main"); got != mainID {
				t.Fatalf("main after batch = %v, want unchanged %v", got, mainID)
			}

			if got := gitResolve(t, repo, "refs/heads/dev"); got != devID {
				t.Fatalf("dev after batch = %v, want unchanged %v", got, devID)
			}
		})
	}
}
