package files_test

import (
	"errors"
	"strings"
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	objectid "lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/ref"
	refstore "lindenii.org/go/furgit/ref/store"
)

func TestFilesTransactionPackedUpdateCreatesLooseOverride(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		_, _, oldID := testRepo.MakeCommit(t, "old packed")
		_, _, newID := testRepo.MakeCommit(t, "new loose")
		testRepo.UpdateRef(t, "refs/heads/main", oldID)
		testRepo.PackRefs(t, "--all", "--prune")

		store := openFilesStore(t, testRepo, algo)

		tx, err := store.BeginTransaction()
		if err != nil {
			t.Fatalf("BeginTransaction: %v", err)
		}

		err = tx.Update("refs/heads/main", newID, oldID)
		if err != nil {
			t.Fatalf("Update queue: %v", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("Commit: %v", err)
		}

		got, err := store.ResolveToDetached("refs/heads/main")
		if err != nil {
			t.Fatalf("ResolveToDetached(main): %v", err)
		}

		if got.ID != newID {
			t.Fatalf("ResolveToDetached(main) = %s, want %s", got.ID, newID)
		}

		packedRefs := string(testRepo.ReadFile(t, "packed-refs"))
		if !strings.Contains(packedRefs, oldID.String()+" refs/heads/main\n") {
			t.Fatalf("packed-refs lost old packed main entry:\n%s", packedRefs)
		}

		looseMain := string(testRepo.ReadFile(t, "refs/heads/main"))
		if strings.TrimSpace(looseMain) != newID.String() {
			t.Fatalf("loose refs/heads/main = %q, want %q", strings.TrimSpace(looseMain), newID.String())
		}
	})
}

func TestFilesTransactionDeletesPackedAndLooseRefs(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		_, _, packedOnlyID := testRepo.MakeCommit(t, "packed only")
		_, _, bothID := testRepo.MakeCommit(t, "both")
		testRepo.UpdateRef(t, "refs/heads/packed", packedOnlyID)
		testRepo.UpdateRef(t, "refs/heads/both", bothID)
		testRepo.PackRefs(t, "--all", "--prune")
		testRepo.UpdateRef(t, "refs/heads/both", bothID)

		store := openFilesStore(t, testRepo, algo)

		tx, err := store.BeginTransaction()
		if err != nil {
			t.Fatalf("BeginTransaction: %v", err)
		}

		err = tx.Delete("refs/heads/packed", packedOnlyID)
		if err != nil {
			t.Fatalf("Delete(packed): %v", err)
		}

		err = tx.Delete("refs/heads/both", bothID)
		if err != nil {
			t.Fatalf("Delete(both): %v", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("Commit(delete): %v", err)
		}

		_, err = store.Resolve("refs/heads/packed")
		if !errors.Is(err, refstore.ErrReferenceNotFound) {
			t.Fatalf("Resolve(packed after delete) error = %v", err)
		}

		_, err = store.Resolve("refs/heads/both")
		if !errors.Is(err, refstore.ErrReferenceNotFound) {
			t.Fatalf("Resolve(both after delete) error = %v", err)
		}

		packedRefs := string(testRepo.ReadFile(t, "packed-refs"))
		if strings.Contains(packedRefs, "refs/heads/packed\n") || strings.Contains(packedRefs, "refs/heads/both\n") {
			t.Fatalf("packed-refs still contains deleted refs:\n%s", packedRefs)
		}
	})
}

func TestFilesTransactionDerefAndDirectSymbolic(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		_, _, firstID := testRepo.MakeCommit(t, "first")
		_, _, secondID := testRepo.MakeCommit(t, "second")
		testRepo.UpdateRef(t, "refs/heads/main", firstID)
		testRepo.SymbolicRef(t, "HEAD", "refs/heads/main")

		store := openFilesStore(t, testRepo, algo)

		tx, err := store.BeginTransaction()
		if err != nil {
			t.Fatalf("BeginTransaction(update): %v", err)
		}

		err = tx.Update("HEAD", secondID, firstID)
		if err != nil {
			t.Fatalf("Update(HEAD): %v", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("Commit(update HEAD): %v", err)
		}

		mainRef, err := store.ResolveToDetached("refs/heads/main")
		if err != nil {
			t.Fatalf("ResolveToDetached(main): %v", err)
		}

		if mainRef.ID != secondID {
			t.Fatalf("main after Update(HEAD) = %s, want %s", mainRef.ID, secondID)
		}

		tx, err = store.BeginTransaction()
		if err != nil {
			t.Fatalf("BeginTransaction(update symbolic): %v", err)
		}

		err = tx.UpdateSymbolic("HEAD", "refs/heads/next", "refs/heads/main")
		if err != nil {
			t.Fatalf("UpdateSymbolic(HEAD): %v", err)
		}

		err = tx.Commit()
		if err != nil {
			t.Fatalf("Commit(update symbolic HEAD): %v", err)
		}

		headRef, err := store.Resolve("HEAD")
		if err != nil {
			t.Fatalf("Resolve(HEAD): %v", err)
		}

		headSym, ok := headRef.(ref.Symbolic)
		if !ok {
			t.Fatalf("Resolve(HEAD) type = %T, want ref.Symbolic", headRef)
		}

		if headSym.Target != "refs/heads/next" {
			t.Fatalf("HEAD target = %q, want %q", headSym.Target, "refs/heads/next")
		}
	})
}
