package files_test

import (
	"errors"
	"os"
	"slices"
	"sync"
	"testing"
	"time"

	"codeberg.org/lindenii/furgit/internal/testgit"
	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/ref/store"
)

func TestFilesTransactionPackedDeleteFailureLeavesRefsUnchanged(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		t.Run("packed-refs.lock held", func(t *testing.T) {
			t.Parallel()

			testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true, RefFormat: "files"})
			_, _, packedID := testRepo.MakeCommit(t, "packed")
			_, _, looseID := testRepo.MakeCommit(t, "loose")
			prefix := "refs/locked-packed-refs"

			testRepo.UpdateRef(t, prefix+"/foo", packedID)
			testRepo.PackRefs(t, "--all", "--prune")
			testRepo.UpdateRef(t, prefix+"/foo", looseID)
			unchanged := forEachRefLines(testRepo.Run(t, "for-each-ref", "--format=%(objectname) %(refname)", prefix))
			testRepo.WriteFile(t, "packed-refs.lock", []byte{}, 0o644)

			store := openFilesStore(t, testRepo, algo)

			tx, err := store.BeginTransaction()
			if err != nil {
				t.Fatalf("BeginTransaction(lock held): %v", err)
			}

			err = tx.Delete(prefix+"/foo", looseID)
			if err != nil {
				t.Fatalf("Delete(lock held) queue: %v", err)
			}

			err = tx.Commit()
			if err == nil {
				t.Fatal("Commit(lock held) unexpectedly succeeded")
			}

			actual := forEachRefLines(testRepo.Run(t, "for-each-ref", "--format=%(objectname) %(refname)", prefix))
			if !slices.Equal(actual, unchanged) {
				t.Fatalf("ShowRef after failed delete = %v, want %v", actual, unchanged)
			}

			got, err := store.ResolveToDetached(prefix + "/foo")
			if err != nil {
				t.Fatalf("ResolveToDetached(lock held): %v", err)
			}

			if got.ID != looseID {
				t.Fatalf("ResolveToDetached(lock held) = %s, want %s", got.ID, looseID)
			}

			gitRoot := testRepo.OpenGitRoot(t)

			_, statErr := gitRoot.Stat(prefix + "/foo.lock")
			if !errors.Is(statErr, os.ErrNotExist) {
				t.Fatalf("unexpected leftover loose lock: %v", statErr)
			}
		})

		t.Run("packed-refs.new exists", func(t *testing.T) {
			t.Parallel()

			testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true, RefFormat: "files"})
			_, _, packedID := testRepo.MakeCommit(t, "packed")
			_, _, looseID := testRepo.MakeCommit(t, "loose")
			prefix := "refs/failed-packed-refs"

			testRepo.UpdateRef(t, prefix+"/foo", packedID)
			testRepo.PackRefs(t, "--all", "--prune")
			testRepo.UpdateRef(t, prefix+"/foo", looseID)
			unchanged := forEachRefLines(testRepo.Run(t, "for-each-ref", "--format=%(objectname) %(refname)", prefix))
			testRepo.WriteFile(t, "packed-refs.new", []byte{}, 0o644)

			store := openFilesStore(t, testRepo, algo)

			tx, err := store.BeginTransaction()
			if err != nil {
				t.Fatalf("BeginTransaction(new exists): %v", err)
			}

			err = tx.Delete(prefix+"/foo", looseID)
			if err != nil {
				t.Fatalf("Delete(new exists) queue: %v", err)
			}

			err = tx.Commit()
			if err == nil {
				t.Fatal("Commit(new exists) unexpectedly succeeded")
			}

			actual := forEachRefLines(testRepo.Run(t, "for-each-ref", "--format=%(objectname) %(refname)", prefix))
			if !slices.Equal(actual, unchanged) {
				t.Fatalf("ShowRef after failed delete = %v, want %v", actual, unchanged)
			}

			got, err := store.ResolveToDetached(prefix + "/foo")
			if err != nil {
				t.Fatalf("ResolveToDetached(new exists): %v", err)
			}

			if got.ID != looseID {
				t.Fatalf("ResolveToDetached(new exists) = %s, want %s", got.ID, looseID)
			}
		})
	})
}

func TestFilesPackedRefDeleteDoesNotCreateDirectories(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true, RefFormat: "files"})
		_, _, commitID := testRepo.MakeCommit(t, "packed-only")
		name := "refs/heads/d1/d2/r1"

		testRepo.UpdateRef(t, name, commitID)
		testRepo.PackRefs(t, "--all", "--prune")

		gitRoot := testRepo.OpenGitRoot(t)

		_, err := gitRoot.Stat("refs/heads/d1/d2")
		if !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("refs/heads/d1/d2 unexpectedly exists before delete: %v", err)
		}

		store := openFilesStore(t, testRepo, algo)

		tx, err := store.BeginTransaction()
		if err != nil {
			t.Fatalf("BeginTransaction: %v", err)
		}

		err = tx.Delete(name, commitID)
		if err != nil {
			t.Fatalf("Delete queue: %v", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("Commit: %v", err)
		}

		_, err = gitRoot.Stat("refs/heads/d1/d2")
		if !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("refs/heads/d1/d2 unexpectedly exists after delete: %v", err)
		}

		_, err = gitRoot.Stat("refs/heads/d1")
		if !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("refs/heads/d1 unexpectedly exists after delete: %v", err)
		}
	})
}

func TestFilesPackedRefIgnoresEmptyDirectories(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true, RefFormat: "files"})
		_, _, commitID := testRepo.MakeCommit(t, "packed-visible")
		prefix := "refs/e-for-each-ref"
		name := prefix + "/foo"

		testRepo.UpdateRef(t, name, commitID)
		expected := forEachRefLines(testRepo.Run(t, "for-each-ref", "--format=%(objectname) %(refname)", prefix))
		testRepo.PackRefs(t, "--all", "--prune")
		testRepo.WriteFileAll(t, prefix+"/foo/bar/baz/.keep", []byte{}, 0o755, 0o644)
		testRepo.Remove(t, prefix+"/foo/bar/baz/.keep")

		store := openFilesStore(t, testRepo, algo)

		got, err := store.ResolveToDetached(name)
		if err != nil {
			t.Fatalf("ResolveToDetached: %v", err)
		}

		if got.ID != commitID {
			t.Fatalf("ResolveToDetached = %s, want %s", got.ID, commitID)
		}

		actual := make([]string, 0)

		listed, err := store.List(prefix + "/*")
		if err != nil {
			t.Fatalf("List: %v", err)
		}

		for _, entry := range listed {
			actual = append(actual, entry.Name())
		}

		fullActual := make([]string, 0, len(actual))
		for _, name := range actual {
			refValue, resolveErr := store.ResolveToDetached(name)
			if resolveErr != nil {
				t.Fatalf("ResolveToDetached(%q): %v", name, resolveErr)
			}

			fullActual = append(fullActual, refValue.ID.String()+" "+name)
		}

		slices.Sort(fullActual)

		if !slices.Equal(fullActual, expected) {
			t.Fatalf("for-each-ref view = %v, want %v", fullActual, expected)
		}
	})
}

func TestFilesDeleteWaitsForPackedRefsLockWithoutIntermediateState(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true, RefFormat: "files"})
		_, _, packedID := testRepo.MakeCommit(t, "packed")
		_, _, looseID := testRepo.MakeCommit(t, "loose")
		prefix := "refs/slow-transaction"

		testRepo.UpdateRef(t, prefix+"/foo", packedID)
		testRepo.PackRefs(t, "--all", "--prune")
		testRepo.UpdateRef(t, prefix+"/foo", looseID)
		testRepo.WriteFile(t, "packed-refs.lock", []byte{}, 0o644)

		store := openFilesStore(t, testRepo, algo)

		tx, err := store.BeginTransaction()
		if err != nil {
			t.Fatalf("BeginTransaction: %v", err)
		}

		err = tx.Delete(prefix+"/foo", looseID)
		if err != nil {
			t.Fatalf("Delete queue: %v", err)
		}

		done := make(chan error, 1)

		var wg sync.WaitGroup

		wg.Go(func() {
			done <- tx.Commit()
		})

		time.Sleep(75 * time.Millisecond)

		select {
		case err := <-done:
			t.Fatalf("Commit finished too early: %v", err)
		default:
		}

		got, err := store.ResolveToDetached(prefix + "/foo")
		if err != nil {
			t.Fatalf("ResolveToDetached while lock held: %v", err)
		}

		if got.ID != looseID {
			t.Fatalf("ResolveToDetached while lock held = %s, want %s", got.ID, looseID)
		}

		testRepo.Remove(t, "packed-refs.lock")

		select {
		case err := <-done:
			if err != nil {
				t.Fatalf("Commit after lock release: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("Commit did not finish after lock release")
		}

		wg.Wait()

		_, err = store.Resolve(prefix + "/foo")
		if !errors.Is(err, refstore.ErrReferenceNotFound) {
			t.Fatalf("Resolve after delete error = %v, want ErrReferenceNotFound", err)
		}
	})
}
