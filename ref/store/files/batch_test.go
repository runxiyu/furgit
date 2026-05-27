package files_test

import (
	"errors"
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	objectid "lindenii.org/go/furgit/object/id"
	refstore "lindenii.org/go/furgit/ref/store"
)

func TestBatchApplyRejectsStaleDeleteAndAppliesIndependentDelete(t *testing.T) {
	t.Parallel()

	//nolint:thelper
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		t.Parallel()

		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo})
		_, _, commitID := testRepo.MakeCommit(t, "base")
		_, _, staleID := testRepo.MakeCommit(t, "stale")
		testRepo.UpdateRef(t, "refs/heads/main", commitID)
		testRepo.UpdateRef(t, "refs/heads/topic", commitID)

		store := openFilesStore(t, testRepo, algo)

		batch, err := store.BeginBatch()
		if err != nil {
			t.Fatalf("BeginBatch: %v", err)
		}

		err = batch.Delete("refs/heads/main", staleID)
		if err != nil {
			t.Fatalf("Delete(main) queue: %v", err)
		}

		err = batch.Delete("refs/heads/topic", commitID)
		if err != nil {
			t.Fatalf("Delete(topic) queue: %v", err)
		}

		results, err := batch.Apply()
		if err != nil {
			t.Fatalf("Apply: %v", err)
		}

		if len(results) != 2 {
			t.Fatalf("len(results) = %d, want 2", len(results))
		}

		if results[0].Status != refstore.BatchStatusRejected {
			t.Fatalf("results[0].Status = %v, want rejected", results[0].Status)
		}

		if _, ok := errors.AsType[*refstore.IncorrectOldValueError](results[0].Error); !errors.Is(results[0].Error, refstore.ErrReferenceNotFound) && !ok {
			t.Fatalf("results[0].Error = %v, want stale-value rejection", results[0].Error)
		}

		if results[1].Status != refstore.BatchStatusApplied {
			t.Fatalf("results[1].Status = %v, want applied", results[1].Status)
		}

		_, err = store.Resolve("refs/heads/main")
		if err != nil {
			t.Fatalf("Resolve(main): %v", err)
		}

		_, err = store.Resolve("refs/heads/topic")
		if err == nil {
			t.Fatal("refs/heads/topic still exists")
		}
	})
}

func TestBatchApplyRejectsDuplicateQueuedRef(t *testing.T) {
	t.Parallel()

	//nolint:thelper
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		t.Parallel()

		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo})
		_, _, commitID := testRepo.MakeCommit(t, "base")
		testRepo.UpdateRef(t, "refs/heads/main", commitID)

		store := openFilesStore(t, testRepo, algo)

		batch, err := store.BeginBatch()
		if err != nil {
			t.Fatalf("BeginBatch: %v", err)
		}

		err = batch.Delete("refs/heads/main", commitID)
		if err != nil {
			t.Fatalf("Delete(main) queue: %v", err)
		}

		err = batch.Verify("refs/heads/main", commitID)
		if err != nil {
			t.Fatalf("Verify(main) queue: %v", err)
		}

		results, err := batch.Apply()
		if err != nil {
			t.Fatalf("Apply: %v", err)
		}

		if len(results) != 2 {
			t.Fatalf("len(results) = %d, want 2", len(results))
		}

		if results[0].Status != refstore.BatchStatusApplied {
			t.Fatalf("results[0].Status = %v, want applied", results[0].Status)
		}

		if results[1].Status != refstore.BatchStatusRejected {
			t.Fatalf("results[1].Status = %v, want rejected", results[1].Status)
		}

		if _, ok := errors.AsType[*refstore.DuplicateUpdateError](results[1].Error); !ok {
			t.Fatalf("results[1].Error = %v, want duplicate update error", results[1].Error)
		}

		_, err = store.Resolve("refs/heads/main")
		if !errors.Is(err, refstore.ErrReferenceNotFound) {
			t.Fatalf("Resolve(main): %v", err)
		}
	})
}
