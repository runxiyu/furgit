package files_test

import (
	"errors"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/ref/store"
)

func TestFilesTransactionDirectSymbolicDeletes(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		_, _, mainID := testRepo.MakeCommit(t, "main")
		testRepo.UpdateRef(t, "refs/heads/main", mainID)

		store := openFilesStore(t, testRepo, algo)

		tx, err := store.BeginTransaction()
		if err != nil {
			t.Fatalf("BeginTransaction(create symref): %v", err)
		}

		err = tx.CreateSymbolic("SYMREF", "refs/heads/main")
		if err != nil {
			t.Fatalf("CreateSymbolic(SYMREF): %v", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("Commit(CreateSymbolic SYMREF): %v", err)
		}

		tx, err = store.BeginTransaction()
		if err != nil {
			t.Fatalf("BeginTransaction(delete symref): %v", err)
		}

		err = tx.DeleteSymbolic("SYMREF", "refs/heads/main")
		if err != nil {
			t.Fatalf("DeleteSymbolic(SYMREF): %v", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("Commit(DeleteSymbolic SYMREF): %v", err)
		}

		_, err = store.Resolve("SYMREF")
		if !errors.Is(err, refstore.ErrReferenceNotFound) {
			t.Fatalf("Resolve(SYMREF after delete) err=%v", err)
		}

		got, err := store.ResolveToDetached("refs/heads/main")
		if err != nil {
			t.Fatalf("ResolveToDetached(main): %v", err)
		}

		if got.ID != mainID {
			t.Fatalf("main after DeleteSymbolic = %s, want %s", got.ID, mainID)
		}
	})
}

func TestFilesTransactionSelfAndDanglingSymrefs(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		_, _, mainID := testRepo.MakeCommit(t, "main")
		testRepo.UpdateRef(t, "refs/heads/main", mainID)

		store := openFilesStore(t, testRepo, algo)

		tx, err := store.BeginTransaction()
		if err != nil {
			t.Fatalf("BeginTransaction(create self): %v", err)
		}

		err = tx.CreateSymbolic("refs/heads/self", "refs/heads/self")
		if err != nil {
			t.Fatalf("CreateSymbolic(self): %v", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("Commit(CreateSymbolic self): %v", err)
		}

		tx, err = store.BeginTransaction()
		if err != nil {
			t.Fatalf("BeginTransaction(delete logical self): %v", err)
		}

		err = tx.Delete("refs/heads/self", mainID)
		if err == nil {
			err = tx.Commit()
		} else {
			_ = tx.Abort()
		}

		if err == nil {
			t.Fatal("Delete(self) unexpectedly succeeded")
		}

		tx, err = store.BeginTransaction()
		if err != nil {
			t.Fatalf("BeginTransaction(delete symbolic self): %v", err)
		}

		err = tx.DeleteSymbolic("refs/heads/self", "refs/heads/self")
		if err != nil {
			t.Fatalf("DeleteSymbolic(self): %v", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("Commit(DeleteSymbolic self): %v", err)
		}

		tx, err = store.BeginTransaction()
		if err != nil {
			t.Fatalf("BeginTransaction(create dangling): %v", err)
		}

		err = tx.CreateSymbolic("refs/heads/dangling", "refs/heads/missing")
		if err != nil {
			t.Fatalf("CreateSymbolic(dangling): %v", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("Commit(CreateSymbolic dangling): %v", err)
		}

		tx, err = store.BeginTransaction()
		if err != nil {
			t.Fatalf("BeginTransaction(delete dangling): %v", err)
		}

		err = tx.DeleteSymbolic("refs/heads/dangling", "refs/heads/missing")
		if err != nil {
			t.Fatalf("DeleteSymbolic(dangling): %v", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("Commit(DeleteSymbolic dangling): %v", err)
		}
	})
}
