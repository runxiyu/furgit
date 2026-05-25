package memory_test

import (
	"errors"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/ref"
	refstore "codeberg.org/lindenii/furgit/ref/store"
	"codeberg.org/lindenii/furgit/ref/store/memory"
)

// Unlike the public ResolveToDetached,
// this one does not resolve symbolic refs.
func resolveDetached(t *testing.T, store *memory.Store, name string) ref.Detached {
	t.Helper()

	resolved, err := store.Resolve(name)
	if err != nil {
		t.Fatalf("Resolve(%q): %v", name, err)
	}

	detached, ok := resolved.(ref.Detached)
	if !ok {
		t.Fatalf("Resolve(%q) = %T, want ref.Detached", name, resolved)
	}

	return detached
}

func TestReadListAndResolveSymbolic(t *testing.T) {
	t.Parallel()

	//nolint:thelper
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		t.Parallel()

		store, err := memory.New(algo)
		if err != nil {
			t.Fatalf("memory.New: %v", err)
		}

		mainID := algo.Sum([]byte("main"))

		tx, err := store.BeginTransaction()
		if err != nil {
			t.Fatalf("BeginTransaction: %v", err)
		}

		err = tx.Create("refs/heads/main", mainID)
		if err != nil {
			t.Fatalf("Create(main): %v", err)
		}

		err = tx.CreateSymbolic("HEAD", "refs/heads/main")
		if err != nil {
			t.Fatalf("CreateSymbolic(HEAD): %v", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("Commit seed refs: %v", err)
		}

		head, err := store.Resolve("HEAD")
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

		detached, err := store.ResolveToDetached("HEAD")
		if err != nil {
			t.Fatalf("ResolveToDetached(HEAD): %v", err)
		}

		if detached.ID != mainID {
			t.Fatalf("ResolveToDetached(HEAD) ID = %v, want %v", detached.ID, mainID)
		}

		listed, err := store.List("")
		if err != nil {
			t.Fatalf("List: %v", err)
		}

		if len(listed) != 2 || listed[0].Name() != "HEAD" || listed[1].Name() != "refs/heads/main" {
			t.Fatalf("List names = %v, want [HEAD refs/heads/main]", listed)
		}
	})
}

func TestTransactionRejectLeavesStoreUnchanged(t *testing.T) {
	t.Parallel()

	//nolint:thelper
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		t.Parallel()

		store, err := memory.New(algo)
		if err != nil {
			t.Fatalf("memory.New: %v", err)
		}

		mainID := algo.Sum([]byte("main"))
		devID := algo.Sum([]byte("dev"))
		nextID := algo.Sum([]byte("next"))
		wrongOld := algo.Sum([]byte("wrong"))

		tx, err := store.BeginTransaction()
		if err != nil {
			t.Fatalf("BeginTransaction: %v", err)
		}

		err = tx.Create("refs/heads/main", mainID)
		if err != nil {
			t.Fatalf("Create(main): %v", err)
		}

		err = tx.Create("refs/heads/dev", devID)
		if err != nil {
			t.Fatalf("Create(dev): %v", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("Commit seed refs: %v", err)
		}

		tx, err = store.BeginTransaction()
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
			t.Fatalf("Commit succeeded, want incorrect old value error")
		}

		var oldValueErr *refstore.IncorrectOldValueError
		if !errors.As(err, &oldValueErr) {
			t.Fatalf("Commit error = %T %v, want IncorrectOldValueError", err, err)
		}

		if got := resolveDetached(t, store, "refs/heads/main").ID; got != mainID {
			t.Fatalf("main after rejected transaction = %v, want %v", got, mainID)
		}

		if got := resolveDetached(t, store, "refs/heads/dev").ID; got != devID {
			t.Fatalf("dev after rejected transaction = %v, want %v", got, devID)
		}
	})
}

func TestBatchRejectsDuplicateResolvedTargetAndAppliesRemainder(t *testing.T) {
	t.Parallel()

	//nolint:thelper
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		t.Parallel()

		store, err := memory.New(algo)
		if err != nil {
			t.Fatalf("memory.New: %v", err)
		}

		mainID := algo.Sum([]byte("main"))
		devID := algo.Sum([]byte("dev"))
		nextMainID := algo.Sum([]byte("next-main"))
		nextDevID := algo.Sum([]byte("next-dev"))
		aliasID := algo.Sum([]byte("alias"))

		tx, err := store.BeginTransaction()
		if err != nil {
			t.Fatalf("BeginTransaction: %v", err)
		}

		err = tx.Create("refs/heads/main", mainID)
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

		err = tx.Commit()
		if err != nil {
			t.Fatalf("Commit seed refs: %v", err)
		}

		batch, err := store.BeginBatch()
		if err != nil {
			t.Fatalf("BeginBatch: %v", err)
		}

		err = batch.Update("refs/heads/main", nextMainID, mainID)
		if err != nil {
			t.Fatalf("Update(main): %v", err)
		}

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

		if results[0].Status != refstore.BatchStatusApplied {
			t.Fatalf("results[0].Status = %v, want applied", results[0].Status)
		}

		if results[1].Status != refstore.BatchStatusRejected {
			t.Fatalf("results[1].Status = %v, want rejected", results[1].Status)
		}

		var duplicateErr *refstore.DuplicateUpdateError
		if !errors.As(results[1].Error, &duplicateErr) {
			t.Fatalf("results[1].Error = %T %v, want DuplicateUpdateError", results[1].Error, results[1].Error)
		}

		if results[2].Status != refstore.BatchStatusApplied {
			t.Fatalf("results[2].Status = %v, want applied", results[2].Status)
		}

		if got := resolveDetached(t, store, "refs/heads/main").ID; got != nextMainID {
			t.Fatalf("main after batch = %v, want %v", got, nextMainID)
		}

		if got := resolveDetached(t, store, "refs/heads/dev").ID; got != nextDevID {
			t.Fatalf("dev after batch = %v, want %v", got, nextDevID)
		}

		resolved, err := store.Resolve("refs/heads/alias")
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
