package loose_test

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
	"codeberg.org/lindenii/furgit/refstore/loose"
)

func openLooseStore(t *testing.T, repoPath string, algo objectid.Algorithm) *loose.Store {
	t.Helper()
	root, err := os.OpenRoot(repoPath)
	if err != nil {
		t.Fatalf("OpenRoot(%q): %v", repoPath, err)
	}
	t.Cleanup(func() { _ = root.Close() })

	store, err := loose.New(root, algo)
	if err != nil {
		t.Fatalf("loose.New: %v", err)
	}
	return store
}

func TestLooseResolveAndResolveFully(t *testing.T) {
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		_, _, commitID := testRepo.MakeCommit(t, "loose refs commit")
		testRepo.UpdateRef(t, "refs/heads/main", commitID)
		testRepo.SymbolicRef(t, "HEAD", "refs/heads/main")

		store := openLooseStore(t, testRepo.Dir(), algo)

		resolvedHead, err := store.Resolve("HEAD")
		if err != nil {
			t.Fatalf("Resolve(HEAD): %v", err)
		}
		headSym, ok := resolvedHead.(ref.Symbolic)
		if !ok {
			t.Fatalf("Resolve(HEAD) type = %T, want ref.Symbolic", resolvedHead)
		}
		if headSym.Target != "refs/heads/main" {
			t.Fatalf("Resolve(HEAD) target = %q, want %q", headSym.Target, "refs/heads/main")
		}

		resolvedMain, err := store.Resolve("refs/heads/main")
		if err != nil {
			t.Fatalf("Resolve(refs/heads/main): %v", err)
		}
		mainDet, ok := resolvedMain.(ref.Detached)
		if !ok {
			t.Fatalf("Resolve(main) type = %T, want ref.Detached", resolvedMain)
		}
		if mainDet.ID != commitID {
			t.Fatalf("Resolve(main) id = %s, want %s", mainDet.ID, commitID)
		}

		fullHead, err := store.ResolveFully("HEAD")
		if err != nil {
			t.Fatalf("ResolveFully(HEAD): %v", err)
		}
		if fullHead.ID != commitID {
			t.Fatalf("ResolveFully(HEAD) id = %s, want %s", fullHead.ID, commitID)
		}

		if _, err := store.Resolve("refs/heads/does-not-exist"); !errors.Is(err, refstore.ErrReferenceNotFound) {
			t.Fatalf("Resolve(not-found) error = %v", err)
		}
	})
}

func TestLooseResolveFullyCycle(t *testing.T) {
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		testRepo.SymbolicRef(t, "refs/heads/a", "refs/heads/b")
		testRepo.SymbolicRef(t, "refs/heads/b", "refs/heads/a")

		store := openLooseStore(t, testRepo.Dir(), algo)
		if _, err := store.ResolveFully("refs/heads/a"); err == nil {
			t.Fatalf("ResolveFully(cycle) expected error")
		}
	})
}

func TestLooseListPattern(t *testing.T) {
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		_, _, commitID := testRepo.MakeCommit(t, "list refs commit")
		testRepo.UpdateRef(t, "refs/heads/main", commitID)
		testRepo.UpdateRef(t, "refs/heads/feature", commitID)
		testRepo.UpdateRef(t, "refs/tags/v1.0.0", commitID)
		testRepo.SymbolicRef(t, "HEAD", "refs/heads/main")

		store := openLooseStore(t, testRepo.Dir(), algo)

		allRefs, err := store.List("")
		if err != nil {
			t.Fatalf("List(\"\"): %v", err)
		}
		allNames := make([]string, 0, len(allRefs))
		for _, entry := range allRefs {
			allNames = append(allNames, entry.Name())
		}
		slices.Sort(allNames)
		wantAll := []string{"HEAD", "refs/heads/feature", "refs/heads/main", "refs/tags/v1.0.0"}
		if !slices.Equal(allNames, wantAll) {
			t.Fatalf("List(\"\") names = %v, want %v", allNames, wantAll)
		}

		headRefs, err := store.List("refs/heads/*")
		if err != nil {
			t.Fatalf("List(refs/heads/*): %v", err)
		}
		headNames := make([]string, 0, len(headRefs))
		for _, entry := range headRefs {
			headNames = append(headNames, entry.Name())
		}
		slices.Sort(headNames)
		wantHeads := []string{"refs/heads/feature", "refs/heads/main"}
		if !slices.Equal(headNames, wantHeads) {
			t.Fatalf("List(refs/heads/*) names = %v, want %v", headNames, wantHeads)
		}
	})
}

func TestLooseMalformedDetachedRef(t *testing.T) {
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		refPath := filepath.Join(testRepo.Dir(), "refs", "heads", "bad")
		if err := os.MkdirAll(filepath.Dir(refPath), 0o755); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		if err := os.WriteFile(refPath, []byte("not-a-hash\n"), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}

		store := openLooseStore(t, testRepo.Dir(), algo)
		if _, err := store.Resolve("refs/heads/bad"); err == nil {
			t.Fatalf("Resolve(malformed) expected error")
		}
	})
}

func TestLooseShorten(t *testing.T) {
	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) {
		testRepo := testgit.NewRepo(t, testgit.RepoOptions{ObjectFormat: algo, Bare: true})
		_, _, commitID := testRepo.MakeCommit(t, "shorten refs commit")
		testRepo.UpdateRef(t, "refs/heads/main", commitID)
		testRepo.UpdateRef(t, "refs/tags/main", commitID)
		testRepo.UpdateRef(t, "refs/remotes/origin/main", commitID)

		store := openLooseStore(t, testRepo.Dir(), algo)

		shortHead, err := store.Shorten("refs/heads/main")
		if err != nil {
			t.Fatalf("Shorten(head): %v", err)
		}
		if shortHead != "heads/main" {
			t.Fatalf("Shorten(refs/heads/main) = %q, want %q", shortHead, "heads/main")
		}

		shortRemote, err := store.Shorten("refs/remotes/origin/main")
		if err != nil {
			t.Fatalf("Shorten(remote): %v", err)
		}
		if shortRemote != "origin/main" {
			t.Fatalf("Shorten(remote) = %q, want %q", shortRemote, "origin/main")
		}

		if _, err := store.Shorten("refs/heads/does-not-exist"); !errors.Is(err, refstore.ErrReferenceNotFound) {
			t.Fatalf("Shorten(not-found) error = %v", err)
		}
	})
}
