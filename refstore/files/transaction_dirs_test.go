package files_test

import (
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/objectid"
)

func TestFilesTransactionEmptyDirectoriesDoNotBlock(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		_, _, oldID := testRepo.MakeCommit(t, "old")
		_, _, newID := testRepo.MakeCommit(t, "new")

		testRepo.UpdateRef(t, "refs/e-verify/foo", oldID)
		testRepo.PackRefs(t, "--all", "--prune")
		testRepo.WriteFileAll(t, "refs/e-verify/foo/bar/baz/.keep", []byte{}, 0o755, 0o644)
		testRepo.Remove(t, "refs/e-verify/foo/bar/baz/.keep")

		store := openFilesStore(t, testRepo, algo)

		tx, err := store.BeginTransaction()
		if err != nil {
			t.Fatalf("BeginTransaction(verify): %v", err)
		}

		err = tx.Verify("refs/e-verify/foo", oldID)
		if err != nil {
			t.Fatalf("Verify with empty directories: %v", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("Commit(verify with empty directories): %v", err)
		}

		testRepo.UpdateRef(t, "refs/e-update/foo", oldID)
		testRepo.PackRefs(t, "--all", "--prune")
		testRepo.WriteFileAll(t, "refs/e-update/foo/bar/baz/.keep", []byte{}, 0o755, 0o644)
		testRepo.Remove(t, "refs/e-update/foo/bar/baz/.keep")

		tx, err = store.BeginTransaction()
		if err != nil {
			t.Fatalf("BeginTransaction(update): %v", err)
		}

		err = tx.Update("refs/e-update/foo", newID, oldID)
		if err != nil {
			t.Fatalf("Update with empty directories: %v", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("Commit(update with empty directories): %v", err)
		}

		got, err := store.ResolveToDetached("refs/e-update/foo")
		if err != nil {
			t.Fatalf("ResolveToDetached(updated foo): %v", err)
		}

		if got.ID != newID {
			t.Fatalf("updated foo = %s, want %s", got.ID, newID)
		}

		testRepo.WriteFileAll(t, "refs/e-create/foo/bar/baz/.keep", []byte{}, 0o755, 0o644)
		testRepo.Remove(t, "refs/e-create/foo/bar/baz/.keep")

		tx, err = store.BeginTransaction()
		if err != nil {
			t.Fatalf("BeginTransaction(create): %v", err)
		}

		err = tx.Create("refs/e-create/foo", oldID)
		if err != nil {
			t.Fatalf("Create with empty directories: %v", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("Commit(create with empty directories): %v", err)
		}

		got, err = store.ResolveToDetached("refs/e-create/foo")
		if err != nil {
			t.Fatalf("ResolveToDetached(created foo): %v", err)
		}

		if got.ID != oldID {
			t.Fatalf("created foo = %s, want %s", got.ID, oldID)
		}

		testRepo.UpdateRef(t, "refs/e-delete/foo", oldID)
		testRepo.PackRefs(t, "--all", "--prune")
		testRepo.WriteFileAll(t, "refs/e-delete/foo/bar/baz/.keep", []byte{}, 0o755, 0o644)
		testRepo.Remove(t, "refs/e-delete/foo/bar/baz/.keep")

		tx, err = store.BeginTransaction()
		if err != nil {
			t.Fatalf("BeginTransaction(delete): %v", err)
		}

		err = tx.Delete("refs/e-delete/foo", oldID)
		if err != nil {
			t.Fatalf("Delete with empty directories: %v", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("Commit(delete with empty directories): %v", err)
		}
	})
}

func TestFilesTransactionNonEmptyDirectoryAndBrokenRefBlockCreate(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		_, _, commitID := testRepo.MakeCommit(t, "base")
		store := openFilesStore(t, testRepo, algo)

		testRepo.WriteFileAll(t, "refs/ne-create/foo/bar/baz.lock", []byte(""), 0o755, 0o644)

		tx, err := store.BeginTransaction()
		if err != nil {
			t.Fatalf("BeginTransaction(non-empty dir): %v", err)
		}

		err = tx.Create("refs/ne-create/foo", commitID)
		if err != nil {
			t.Fatalf("Create(non-empty dir) queue: %v", err)
		}

		err = tx.Commit()
		if err == nil {
			t.Fatal("Commit(non-empty dir) unexpectedly succeeded")
		}

		testRepo.WriteFileAll(t, "refs/broken/foo", []byte("gobbledigook\n"), 0o755, 0o644)

		tx, err = store.BeginTransaction()
		if err != nil {
			t.Fatalf("BeginTransaction(broken ref): %v", err)
		}

		err = tx.Create("refs/broken/foo", commitID)
		if err != nil {
			t.Fatalf("Create(broken ref) queue: %v", err)
		}

		err = tx.Commit()
		if err == nil {
			t.Fatal("Commit(broken ref) unexpectedly succeeded")
		}
	})
}

func TestFilesTransactionIndirectCreateMatchesGit(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		t.Run("non-empty directory blocks", func(t *testing.T) {
			t.Parallel()

			repo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, RefFormat: "files"})
			_, _, innerID := repo.MakeCommit(t, "inner")
			prefix := "refs/ne-indirect-create"

			repo.SymbolicRef(t, prefix+"/symref", prefix+"/foo")
			repo.WriteFileAll(t, ".git/"+prefix+"/foo/bar/baz.lock", []byte{}, 0o755, 0o644)
			store := openFilesStore(t, repo, algo)

			tx, err := store.BeginTransaction()
			if err != nil {
				t.Fatalf("BeginTransaction(non-empty): %v", err)
			}

			err = tx.Create(prefix+"/symref", innerID)
			if err != nil {
				t.Fatalf("Create(non-empty) queue: %v", err)
			}

			err = tx.Commit()
			if err == nil {
				t.Fatal("Commit(non-empty) unexpectedly succeeded")
			}
		})

		t.Run("broken referent blocks", func(t *testing.T) {
			t.Parallel()

			repo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, RefFormat: "files"})
			_, _, commitID := repo.MakeCommit(t, "broken")
			prefix := "refs/broken-indirect-create"

			repo.SymbolicRef(t, prefix+"/symref", prefix+"/foo")
			repo.WriteFileAll(t, ".git/"+prefix+"/foo", []byte("gobbledigook\n"), 0o755, 0o644)
			store := openFilesStore(t, repo, algo)

			tx, err := store.BeginTransaction()
			if err != nil {
				t.Fatalf("BeginTransaction(broken): %v", err)
			}

			err = tx.Create(prefix+"/symref", commitID)
			if err != nil {
				t.Fatalf("Create(broken) queue: %v", err)
			}

			err = tx.Commit()
			if err == nil {
				t.Fatal("Commit(broken) unexpectedly succeeded")
			}
		})
	})
}
