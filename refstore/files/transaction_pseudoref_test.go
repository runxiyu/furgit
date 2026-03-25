package files_test

import (
	"errors"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/refstore"
)

func TestFilesTransactionPseudorefLifecycle(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		_, _, aID := testRepo.MakeCommit(t, "A")
		_, _, bID := testRepo.MakeCommit(t, "B")
		_, _, cID := testRepo.MakeCommit(t, "C")

		store := openFilesStore(t, testRepo, algo)

		tx, err := store.BeginTransaction()
		if err != nil {
			t.Fatalf("BeginTransaction(create): %v", err)
		}

		err = tx.Create("PSEUDOREF", aID)
		if err != nil {
			t.Fatalf("Create(PSEUDOREF): %v", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("Commit(create PSEUDOREF): %v", err)
		}

		got, err := store.ResolveToDetached("PSEUDOREF")
		if err != nil {
			t.Fatalf("ResolveToDetached(PSEUDOREF): %v", err)
		}

		if got.ID != aID {
			t.Fatalf("PSEUDOREF after create = %s, want %s", got.ID, aID)
		}

		tx, err = store.BeginTransaction()
		if err != nil {
			t.Fatalf("BeginTransaction(update): %v", err)
		}

		err = tx.Update("PSEUDOREF", bID, aID)
		if err != nil {
			t.Fatalf("Update(PSEUDOREF): %v", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("Commit(update PSEUDOREF): %v", err)
		}

		got, err = store.ResolveToDetached("PSEUDOREF")
		if err != nil {
			t.Fatalf("ResolveToDetached(PSEUDOREF) after update: %v", err)
		}

		if got.ID != bID {
			t.Fatalf("PSEUDOREF after update = %s, want %s", got.ID, bID)
		}

		tx, err = store.BeginTransaction()
		if err != nil {
			t.Fatalf("BeginTransaction(stale update): %v", err)
		}

		err = tx.Update("PSEUDOREF", cID, aID)
		if err != nil {
			t.Fatalf("queue stale update: %v", err)
		}

		err = tx.Commit()
		if err == nil {
			t.Fatal("stale pseudoref update unexpectedly succeeded")
		}

		tx, err = store.BeginTransaction()
		if err != nil {
			t.Fatalf("BeginTransaction(delete): %v", err)
		}

		err = tx.Delete("PSEUDOREF", bID)
		if err != nil {
			t.Fatalf("Delete(PSEUDOREF): %v", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("Commit(delete PSEUDOREF): %v", err)
		}

		_, err = store.Resolve("PSEUDOREF")
		if !errors.Is(err, refstore.ErrReferenceNotFound) {
			t.Fatalf("Resolve(PSEUDOREF after delete) err=%v", err)
		}
	})
}
