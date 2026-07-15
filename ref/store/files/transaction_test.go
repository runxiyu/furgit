package files_test

import (
	"errors"
	"os"
	"testing"
	"time"

	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/ref"
	"lindenii.org/go/furgit/ref/store"
	"lindenii.org/go/furgit/ref/store/files"
)

// commitOps runs fn against a fresh transaction and commits it,
// failing the test on any error.
func commitOps(t *testing.T, filesStore *files.Files, fn func(tx store.Transaction)) {
	t.Helper()

	tx, err := filesStore.BeginTransaction()
	if err != nil {
		t.Fatalf("BeginTransaction: %v", err)
	}

	fn(tx)

	err = tx.Commit()
	if err != nil {
		t.Fatalf("Commit: %v", err)
	}
}

func gitResolve(t *testing.T, repo *testgit.Repo, name string) id.ObjectID {
	t.Helper()

	oid, err := repo.RevParse(t, name)
	if err != nil {
		t.Fatalf("RevParse(%q): %v", name, err)
	}

	return oid
}

func TestTransactionCreateUpdateDelete(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo := newFilesRepo(t, objectFormat)
			firstID := makeCommit(t, repo, "first")
			secondID := makeCommit(t, repo, "second")
			filesStore := openStore(t, repo, objectFormat)

			commitOps(t, filesStore, func(tx store.Transaction) {
				err := tx.Create("refs/heads/topic", firstID)
				if err != nil {
					t.Fatalf("Create: %v", err)
				}
			})

			if got := gitResolve(t, repo, "refs/heads/topic"); got != firstID {
				t.Fatalf("git sees %v, want %v", got, firstID)
			}

			commitOps(t, filesStore, func(tx store.Transaction) {
				err := tx.Update("refs/heads/topic", secondID, firstID)
				if err != nil {
					t.Fatalf("Update: %v", err)
				}
			})

			if got := gitResolve(t, repo, "refs/heads/topic"); got != secondID {
				t.Fatalf("git sees %v, want %v", got, secondID)
			}

			commitOps(t, filesStore, func(tx store.Transaction) {
				err := tx.Delete("refs/heads/topic", secondID)
				if err != nil {
					t.Fatalf("Delete: %v", err)
				}
			})

			_, err := filesStore.Resolve("refs/heads/topic")
			if !errors.Is(err, store.ErrReferenceNotFound) {
				t.Fatalf("Resolve after delete = %v, want ErrReferenceNotFound", err)
			}
		})
	}
}

func TestTransactionDeletePacked(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo := newFilesRepo(t, objectFormat)
			commitID := makeCommit(t, repo, "c1")
			keptID := makeCommit(t, repo, "kept")

			err := repo.UpdateRef(t, "refs/heads/doomed", commitID)
			if err != nil {
				t.Fatalf("UpdateRef(doomed): %v", err)
			}

			err = repo.UpdateRef(t, "refs/heads/kept", keptID)
			if err != nil {
				t.Fatalf("UpdateRef(kept): %v", err)
			}

			err = repo.PackRefs(t)
			if err != nil {
				t.Fatalf("PackRefs: %v", err)
			}

			filesStore := openStore(t, repo, objectFormat)

			commitOps(t, filesStore, func(tx store.Transaction) {
				err := tx.Delete("refs/heads/doomed", commitID)
				if err != nil {
					t.Fatalf("Delete: %v", err)
				}
			})

			_, err = filesStore.Resolve("refs/heads/doomed")
			if !errors.Is(err, store.ErrReferenceNotFound) {
				t.Fatalf("Resolve(doomed) = %v, want ErrReferenceNotFound", err)
			}

			if got := gitResolve(t, repo, "refs/heads/kept"); got != keptID {
				t.Fatalf("git sees kept = %v, want %v", got, keptID)
			}

			gitdir := openGitDirRoot(t, repo)

			_, err = gitdir.Stat("packed-refs.lock")
			if !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("leftover packed-refs.lock: %v", err)
			}
		})
	}
}

func TestTransactionDeleteLooseAndPacked(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo := newFilesRepo(t, objectFormat)
			packedID := makeCommit(t, repo, "packed")
			looseID := makeCommit(t, repo, "loose")

			err := repo.UpdateRef(t, "refs/heads/both", packedID)
			if err != nil {
				t.Fatalf("UpdateRef(packed): %v", err)
			}

			err = repo.PackRefs(t)
			if err != nil {
				t.Fatalf("PackRefs: %v", err)
			}

			err = repo.UpdateRef(t, "refs/heads/both", looseID)
			if err != nil {
				t.Fatalf("UpdateRef(loose): %v", err)
			}

			filesStore := openStore(t, repo, objectFormat)

			commitOps(t, filesStore, func(tx store.Transaction) {
				err := tx.Delete("refs/heads/both", looseID)
				if err != nil {
					t.Fatalf("Delete: %v", err)
				}
			})

			_, err = filesStore.Resolve("refs/heads/both")
			if !errors.Is(err, store.ErrReferenceNotFound) {
				t.Fatalf("Resolve = %v, want ErrReferenceNotFound; packed version must not resurface", err)
			}
		})
	}
}

func TestTransactionSymbolicOps(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo := newFilesRepo(t, objectFormat)
			commitID := makeCommit(t, repo, "c1")

			err := repo.UpdateRef(t, "refs/heads/main", commitID)
			if err != nil {
				t.Fatalf("UpdateRef: %v", err)
			}

			err = repo.UpdateRef(t, "refs/heads/other", commitID)
			if err != nil {
				t.Fatalf("UpdateRef(other): %v", err)
			}

			filesStore := openStore(t, repo, objectFormat)

			commitOps(t, filesStore, func(tx store.Transaction) {
				err := tx.CreateSymbolic("refs/sym/head", "refs/heads/main")
				if err != nil {
					t.Fatalf("CreateSymbolic: %v", err)
				}
			})

			if got := gitResolve(t, repo, "refs/sym/head"); got != commitID {
				t.Fatalf("git resolves symref to %v, want %v", got, commitID)
			}

			commitOps(t, filesStore, func(tx store.Transaction) {
				err := tx.UpdateSymbolic("refs/sym/head", "refs/heads/other", "refs/heads/main")
				if err != nil {
					t.Fatalf("UpdateSymbolic: %v", err)
				}
			})

			resolved, err := filesStore.Resolve("refs/sym/head")
			if err != nil {
				t.Fatalf("Resolve: %v", err)
			}

			symbolic, ok := resolved.(ref.Symbolic)
			if !ok || symbolic.Target != "refs/heads/other" {
				t.Fatalf("Resolve = %#v, want symbolic to refs/heads/other", resolved)
			}

			commitOps(t, filesStore, func(tx store.Transaction) {
				err := tx.DeleteSymbolic("refs/sym/head", "refs/heads/other")
				if err != nil {
					t.Fatalf("DeleteSymbolic: %v", err)
				}
			})

			_, err = filesStore.Resolve("refs/sym/head")
			if !errors.Is(err, store.ErrReferenceNotFound) {
				t.Fatalf("Resolve after delete = %v, want ErrReferenceNotFound", err)
			}
		})
	}
}

func TestTransactionHeadDereference(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo := newFilesRepo(t, objectFormat)
			firstID := makeCommit(t, repo, "first")
			secondID := makeCommit(t, repo, "second")

			err := repo.UpdateRef(t, "refs/heads/main", firstID)
			if err != nil {
				t.Fatalf("UpdateRef: %v", err)
			}

			err = repo.SymbolicRefUpdate(t, "HEAD", "refs/heads/main")
			if err != nil {
				t.Fatalf("SymbolicRefUpdate: %v", err)
			}

			filesStore := openStore(t, repo, objectFormat)

			commitOps(t, filesStore, func(tx store.Transaction) {
				err := tx.Update("HEAD", secondID, firstID)
				if err != nil {
					t.Fatalf("Update(HEAD): %v", err)
				}
			})

			if got := gitResolve(t, repo, "refs/heads/main"); got != secondID {
				t.Fatalf("git sees branch %v, want %v", got, secondID)
			}

			resolved, err := filesStore.Resolve("HEAD")
			if err != nil {
				t.Fatalf("Resolve(HEAD): %v", err)
			}

			if _, ok := resolved.(ref.Symbolic); !ok {
				t.Fatalf("HEAD = %T, want still symbolic", resolved)
			}
		})
	}
}

func TestTransactionExpectedValueFailures(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo := newFilesRepo(t, objectFormat)
			commitID := makeCommit(t, repo, "c1")
			wrongID := makeCommit(t, repo, "wrong")

			err := repo.UpdateRef(t, "refs/heads/main", commitID)
			if err != nil {
				t.Fatalf("UpdateRef: %v", err)
			}

			filesStore := openStore(t, repo, objectFormat)

			tx, err := filesStore.BeginTransaction()
			if err != nil {
				t.Fatalf("BeginTransaction: %v", err)
			}

			err = tx.Update("refs/heads/main", wrongID, wrongID)
			if err != nil {
				t.Fatalf("Update queue: %v", err)
			}

			err = tx.Commit()

			wrongOld, ok := errors.AsType[*store.WrongOldIDError](err)
			if !ok {
				t.Fatalf("Commit err = %v, want WrongOldIDError", err)
			}

			if wrongOld.Actual != commitID {
				t.Fatalf("Actual = %v, want %v", wrongOld.Actual, commitID)
			}

			if got := gitResolve(t, repo, "refs/heads/main"); got != commitID {
				t.Fatalf("git sees %v, want unchanged %v", got, commitID)
			}

			tx, err = filesStore.BeginTransaction()
			if err != nil {
				t.Fatalf("BeginTransaction: %v", err)
			}

			err = tx.Create("refs/heads/main", wrongID)
			if err != nil {
				t.Fatalf("Create queue: %v", err)
			}

			err = tx.Commit()
			if !errors.Is(err, store.ErrCreateExists) {
				t.Fatalf("Commit err = %v, want ErrCreateExists", err)
			}
		})
	}
}

func TestTransactionVerify(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo := newFilesRepo(t, objectFormat)
			commitID := makeCommit(t, repo, "c1")

			err := repo.UpdateRef(t, "refs/heads/main", commitID)
			if err != nil {
				t.Fatalf("UpdateRef: %v", err)
			}

			err = repo.SymbolicRefUpdate(t, "HEAD", "refs/heads/main")
			if err != nil {
				t.Fatalf("SymbolicRefUpdate: %v", err)
			}

			filesStore := openStore(t, repo, objectFormat)

			commitOps(t, filesStore, func(tx store.Transaction) {
				err := tx.Verify("refs/heads/main", commitID)
				if err != nil {
					t.Fatalf("Verify queue: %v", err)
				}

				err = tx.VerifySymbolic("HEAD", "refs/heads/main")
				if err != nil {
					t.Fatalf("VerifySymbolic queue: %v", err)
				}
			})
		})
	}
}

func TestTransactionDuplicateUpdate(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo := newFilesRepo(t, objectFormat)
			commitID := makeCommit(t, repo, "c1")
			filesStore := openStore(t, repo, objectFormat)

			tx, err := filesStore.BeginTransaction()
			if err != nil {
				t.Fatalf("BeginTransaction: %v", err)
			}

			err = tx.Create("refs/heads/dup", commitID)
			if err != nil {
				t.Fatalf("Create queue: %v", err)
			}

			err = tx.Create("refs/heads/dup", commitID)
			if err != nil {
				t.Fatalf("Create queue 2: %v", err)
			}

			err = tx.Commit()
			if !errors.Is(err, store.ErrDuplicateUpdate) {
				t.Fatalf("Commit err = %v, want ErrDuplicateUpdate", err)
			}
		})
	}
}

func TestTransactionNameConflicts(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo := newFilesRepo(t, objectFormat)
			commitID := makeCommit(t, repo, "c1")

			err := repo.UpdateRef(t, "refs/heads/a", commitID)
			if err != nil {
				t.Fatalf("UpdateRef: %v", err)
			}

			filesStore := openStore(t, repo, objectFormat)

			tx, err := filesStore.BeginTransaction()
			if err != nil {
				t.Fatalf("BeginTransaction: %v", err)
			}

			err = tx.Create("refs/heads/a/b", commitID)
			if err != nil {
				t.Fatalf("Create queue: %v", err)
			}

			err = tx.Commit()

			conflict, ok := errors.AsType[*store.NameConflictError](err)
			if !ok {
				t.Fatalf("Commit err = %v, want NameConflictError", err)
			}

			if conflict.Other != "refs/heads/a" {
				t.Fatalf("Other = %q, want refs/heads/a", conflict.Other)
			}
		})
	}
}

func TestTransactionEmptyDirRemovedOnCreate(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo := newFilesRepo(t, objectFormat)
			commitID := makeCommit(t, repo, "c1")
			gitdir := openGitDirRoot(t, repo)

			err := gitdir.MkdirAll("refs/heads/blocked/leftover", 0o755)
			if err != nil {
				t.Fatalf("MkdirAll: %v", err)
			}

			filesStore := openStore(t, repo, objectFormat)

			commitOps(t, filesStore, func(tx store.Transaction) {
				err := tx.Create("refs/heads/blocked", commitID)
				if err != nil {
					t.Fatalf("Create: %v", err)
				}
			})

			if got := gitResolve(t, repo, "refs/heads/blocked"); got != commitID {
				t.Fatalf("git sees %v, want %v", got, commitID)
			}
		})
	}
}

func TestTransactionDeletePrunesEmptyParents(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo := newFilesRepo(t, objectFormat)
			commitID := makeCommit(t, repo, "c1")

			err := repo.UpdateRef(t, "refs/heads/d1/d2/leaf", commitID)
			if err != nil {
				t.Fatalf("UpdateRef: %v", err)
			}

			filesStore := openStore(t, repo, objectFormat)

			commitOps(t, filesStore, func(tx store.Transaction) {
				err := tx.Delete("refs/heads/d1/d2/leaf", commitID)
				if err != nil {
					t.Fatalf("Delete: %v", err)
				}
			})

			gitdir := openGitDirRoot(t, repo)

			_, err = gitdir.Stat("refs/heads/d1")
			if !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("refs/heads/d1 still exists: %v", err)
			}

			_, err = gitdir.Stat("refs/heads")
			if err != nil {
				t.Fatalf("refs/heads must be preserved: %v", err)
			}
		})
	}
}

func TestTransactionDeleteRemovesReflog(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo := newFilesRepo(t, objectFormat)
			commitID := makeCommit(t, repo, "c1")

			err := repo.UpdateRef(t, "refs/heads/logged", commitID)
			if err != nil {
				t.Fatalf("UpdateRef: %v", err)
			}

			gitdir := openGitDirRoot(t, repo)

			_, err = gitdir.Stat("logs/refs/heads/logged")
			if err != nil {
				t.Fatalf("git did not write a reflog: %v", err)
			}

			filesStore := openStore(t, repo, objectFormat)

			commitOps(t, filesStore, func(tx store.Transaction) {
				err := tx.Delete("refs/heads/logged", commitID)
				if err != nil {
					t.Fatalf("Delete: %v", err)
				}
			})

			_, err = gitdir.Stat("logs/refs/heads/logged")
			if !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("reflog still exists after delete: %v", err)
			}
		})
	}
}

func TestTransactionPackedLockHeldFailsUnchanged(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo := newFilesRepo(t, objectFormat)
			commitID := makeCommit(t, repo, "c1")

			err := repo.UpdateRef(t, "refs/heads/held", commitID)
			if err != nil {
				t.Fatalf("UpdateRef: %v", err)
			}

			err = repo.PackRefs(t)
			if err != nil {
				t.Fatalf("PackRefs: %v", err)
			}

			gitdir := openGitDirRoot(t, repo)

			err = gitdir.WriteFile("packed-refs.lock", nil, 0o644)
			if err != nil {
				t.Fatalf("WriteFile(lock): %v", err)
			}

			filesStore := openStoreOptions(t, repo, objectFormat, files.Options{
				LooseLockTimeout:  0,
				PackedLockTimeout: 0,
			})

			tx, err := filesStore.BeginTransaction()
			if err != nil {
				t.Fatalf("BeginTransaction: %v", err)
			}

			err = tx.Delete("refs/heads/held", commitID)
			if err != nil {
				t.Fatalf("Delete queue: %v", err)
			}

			err = tx.Commit()
			if err == nil {
				t.Fatal("Commit unexpectedly succeeded")
			}

			direct, err := filesStore.ResolveToDirect("refs/heads/held")
			if err != nil {
				t.Fatalf("ResolveToDirect after failure: %v", err)
			}

			if direct.ID != commitID {
				t.Fatalf("ID = %v, want unchanged %v", direct.ID, commitID)
			}

			_, err = gitdir.Stat("refs/heads/held.lock")
			if !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("leftover loose lock: %v", err)
			}

			_, err = gitdir.Stat("packed-refs.lock")
			if err != nil {
				t.Fatalf("foreign packed-refs.lock must not be removed: %v", err)
			}
		})
	}
}

func TestTransactionWaitsForPackedLock(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo := newFilesRepo(t, objectFormat)
			commitID := makeCommit(t, repo, "c1")

			err := repo.UpdateRef(t, "refs/heads/waited", commitID)
			if err != nil {
				t.Fatalf("UpdateRef: %v", err)
			}

			err = repo.PackRefs(t)
			if err != nil {
				t.Fatalf("PackRefs: %v", err)
			}

			gitdir := openGitDirRoot(t, repo)

			err = gitdir.WriteFile("packed-refs.lock", nil, 0o644)
			if err != nil {
				t.Fatalf("WriteFile(lock): %v", err)
			}

			filesStore := openStoreOptions(t, repo, objectFormat, files.Options{
				LooseLockTimeout:  files.DefaultLooseLockTimeout,
				PackedLockTimeout: -1,
			})

			tx, err := filesStore.BeginTransaction()
			if err != nil {
				t.Fatalf("BeginTransaction: %v", err)
			}

			err = tx.Delete("refs/heads/waited", commitID)
			if err != nil {
				t.Fatalf("Delete queue: %v", err)
			}

			done := make(chan error, 1)

			go func() {
				done <- tx.Commit()
			}()

			time.Sleep(75 * time.Millisecond)

			select {
			case commitErr := <-done:
				t.Fatalf("Commit finished while lock held: %v", commitErr)
			default:
			}

			err = gitdir.Remove("packed-refs.lock")
			if err != nil {
				t.Fatalf("Remove(lock): %v", err)
			}

			select {
			case commitErr := <-done:
				if commitErr != nil {
					t.Fatalf("Commit after lock release: %v", commitErr)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("Commit did not finish after lock release")
			}

			_, err = filesStore.Resolve("refs/heads/waited")
			if !errors.Is(err, store.ErrReferenceNotFound) {
				t.Fatalf("Resolve after delete = %v, want ErrReferenceNotFound", err)
			}
		})
	}
}
