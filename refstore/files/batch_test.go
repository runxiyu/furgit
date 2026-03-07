package files_test

import (
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/objectid"
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

		batch.Delete("refs/heads/main", staleID)
		batch.Delete("refs/heads/topic", commitID)

		results, err := batch.Apply()
		if err != nil {
			t.Fatalf("Apply: %v", err)
		}

		if len(results) != 2 {
			t.Fatalf("len(results) = %d, want 2", len(results))
		}

		if results[0].Error == nil {
			t.Fatal("stale delete unexpectedly succeeded")
		}

		if results[1].Error != nil {
			t.Fatalf("valid delete failed: %v", results[1].Error)
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

		batch.Delete("refs/heads/main", commitID)
		batch.Verify("refs/heads/main", commitID)

		results, err := batch.Apply()
		if err == nil {
			t.Fatal("Apply unexpectedly succeeded")
		}

		if len(results) != 2 {
			t.Fatalf("len(results) = %d, want 2", len(results))
		}

		if results[1].Error == nil {
			t.Fatal("duplicate ref operation did not report an error")
		}

		_, err = store.Resolve("refs/heads/main")
		if err != nil {
			t.Fatalf("Resolve(main): %v", err)
		}
	})
}
