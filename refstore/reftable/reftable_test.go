package reftable_test

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/ref"
	"codeberg.org/lindenii/furgit/refstore"
	"codeberg.org/lindenii/furgit/refstore/reftable"
)

// newBareReftableRepo creates a bare repository that uses reftable ref storage.
func newBareReftableRepo(tb testing.TB, algo objectid.Algorithm) *testgit.TestRepo {
	tb.Helper()

	return testgit.NewRepo(tb, testgit.RepoOptions{
		ObjectFormat: algo,
		Bare:         true,
		RefFormat:    "reftable",
	})
}

// openStore opens a reftable store against repoDir/reftable.
func openStore(tb testing.TB, repoDir string, algo objectid.Algorithm) *reftable.Store {
	tb.Helper()

	root, err := os.OpenRoot(filepath.Join(repoDir, "reftable"))
	if err != nil {
		tb.Fatalf("OpenRoot(reftable): %v", err)
	}

	tb.Cleanup(func() { _ = root.Close() })

	store, err := reftable.New(root, algo)
	if err != nil {
		tb.Fatalf("reftable.New: %v", err)
	}

	return store
}

func TestResolveAndResolveFully(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		repo := newBareReftableRepo(t, algo)
		_, _, id := repo.MakeCommit(t, "resolve")
		repo.UpdateRef(t, "refs/heads/main", id)
		repo.SymbolicRef(t, "HEAD", "refs/heads/main")

		store := openStore(t, repo.Dir(), algo)

		head, err := store.Resolve("HEAD")
		if err != nil {
			t.Fatalf("Resolve(HEAD): %v", err)
		}

		sym, ok := head.(ref.Symbolic)
		if !ok {
			t.Fatalf("Resolve(HEAD) type = %T, want ref.Symbolic", head)
		}

		if sym.Target != "refs/heads/main" {
			t.Fatalf("Resolve(HEAD) target = %q, want refs/heads/main", sym.Target)
		}

		main, err := store.ResolveFully("HEAD")
		if err != nil {
			t.Fatalf("ResolveFully(HEAD): %v", err)
		}

		if main.ID != id {
			t.Fatalf("ResolveFully(HEAD) id = %s, want %s", main.ID, id)
		}

		_, err = store.Resolve("refs/heads/missing")
		if !errors.Is(err, refstore.ErrReferenceNotFound) {
			t.Fatalf("Resolve(missing) error = %v", err)
		}
	})
}

func TestResolveFullyCycle(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		repo := newBareReftableRepo(t, algo)
		repo.SymbolicRef(t, "refs/heads/a", "refs/heads/b")
		repo.SymbolicRef(t, "refs/heads/b", "refs/heads/a")

		store := openStore(t, repo.Dir(), algo)

		_, err := store.ResolveFully("refs/heads/a")
		if err == nil {
			t.Fatalf("ResolveFully(cycle) expected error")
		}
	})
}

func TestListAndShorten(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		repo := newBareReftableRepo(t, algo)
		_, _, id := repo.MakeCommit(t, "list")
		repo.UpdateRef(t, "refs/heads/main", id)
		repo.UpdateRef(t, "refs/heads/feature", id)
		repo.UpdateRef(t, "refs/tags/main", id)
		repo.UpdateRef(t, "refs/remotes/origin/main", id)

		store := openStore(t, repo.Dir(), algo)

		all, err := store.List("")
		if err != nil {
			t.Fatalf("List(all): %v", err)
		}

		names := make([]string, 0, len(all))
		for _, entry := range all {
			names = append(names, entry.Name())
		}

		want := []string{"HEAD", "refs/heads/feature", "refs/heads/main", "refs/remotes/origin/main", "refs/tags/main"}
		if !slices.Equal(names, want) {
			t.Fatalf("List(all) = %v, want %v", names, want)
		}

		heads, err := store.List("refs/heads/*")
		if err != nil {
			t.Fatalf("List(heads): %v", err)
		}

		headNames := make([]string, 0, len(heads))
		for _, entry := range heads {
			headNames = append(headNames, entry.Name())
		}

		wantHeads := []string{"refs/heads/feature", "refs/heads/main"}
		if !slices.Equal(headNames, wantHeads) {
			t.Fatalf("List(heads) = %v, want %v", headNames, wantHeads)
		}

		short, err := store.Shorten("refs/remotes/origin/main")
		if err != nil {
			t.Fatalf("Shorten(remote): %v", err)
		}

		if short != "origin/main" {
			t.Fatalf("Shorten(remote) = %q, want origin/main", short)
		}
	})
}

func TestTombstoneNewestWins(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		repo := newBareReftableRepo(t, algo)
		_, _, oldID := repo.MakeCommit(t, "old")
		repo.UpdateRef(t, "refs/heads/main", oldID)
		_, _, newID := repo.MakeCommit(t, "new")
		repo.UpdateRef(t, "refs/heads/main", newID)
		repo.DeleteRef(t, "refs/heads/main")

		store := openStore(t, repo.Dir(), algo)

		_, err := store.Resolve("refs/heads/main")
		if !errors.Is(err, refstore.ErrReferenceNotFound) {
			t.Fatalf("Resolve(main) after delete error = %v", err)
		}
	})
}

func TestAnnotatedTagPeeled(t *testing.T) {
	t.Parallel()
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		repo := newBareReftableRepo(t, algo)
		_, _, commitID := repo.MakeCommit(t, "tagged")
		tagID := repo.TagAnnotated(t, "v1.0.0", commitID, "annotated")

		store := openStore(t, repo.Dir(), algo)

		resolved, err := store.Resolve("refs/tags/v1.0.0")
		if err != nil {
			t.Fatalf("Resolve(tag): %v", err)
		}

		detached, ok := resolved.(ref.Detached)
		if !ok {
			t.Fatalf("Resolve(tag) type = %T, want ref.Detached", resolved)
		}

		if detached.ID != tagID {
			t.Fatalf("Resolve(tag) id = %s, want %s", detached.ID, tagID)
		}

		if detached.Peeled == nil {
			t.Fatalf("Resolve(tag) peeled = nil")
		}

		if *detached.Peeled != commitID {
			t.Fatalf("Resolve(tag) peeled = %s, want %s", *detached.Peeled, commitID)
		}
	})
}
