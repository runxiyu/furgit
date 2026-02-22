package repository_test

import (
	"os"
	"strings"
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/ref"
	"codeberg.org/lindenii/furgit/repository"
)

func TestRefConvenienceMethods(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		repoHarness := testgit.NewRepo(t, testgit.RepoOptions{
			ObjectFormat: algo,
			Bare:         true,
			RefFormat:    "files",
		})

		_, _, commitID := repoHarness.MakeCommit(t, "refs wrappers")
		repoHarness.UpdateRef(t, "refs/heads/main", commitID)
		repoHarness.SymbolicRef(t, "HEAD", "refs/heads/main")
		repoHarness.UpdateRef(t, "refs/tags/v1", commitID)

		root, err := os.OpenRoot(repoHarness.Dir())
		if err != nil {
			t.Fatalf("os.OpenRoot: %v", err)
		}
		defer func() { _ = root.Close() }()

		repo, err := repository.Open(root)
		if err != nil {
			t.Fatalf("repository.Open: %v", err)
		}
		defer func() { _ = repo.Close() }()

		resolved, err := repo.ResolveRef("HEAD")
		if err != nil {
			t.Fatalf("ResolveRef(HEAD): %v", err)
		}
		sym, ok := resolved.(ref.Symbolic)
		if !ok {
			t.Fatalf("ResolveRef(HEAD) type = %T, want ref.Symbolic", resolved)
		}
		if sym.Target != "refs/heads/main" {
			t.Fatalf("ResolveRef(HEAD) target = %q, want %q", sym.Target, "refs/heads/main")
		}

		fully, err := repo.ResolveRefFully("HEAD")
		if err != nil {
			t.Fatalf("ResolveRefFully(HEAD): %v", err)
		}
		if fully.ID != commitID {
			t.Fatalf("ResolveRefFully(HEAD) id = %s, want %s", fully.ID, commitID)
		}

		refs, err := repo.ListRefs("refs/*/*")
		if err != nil {
			t.Fatalf("ListRefs: %v", err)
		}
		if len(refs) < 2 {
			t.Fatalf("ListRefs returned %d refs, want >= 2", len(refs))
		}

		short, err := repo.ShortenRef("refs/heads/main")
		if err != nil {
			t.Fatalf("ShortenRef: %v", err)
		}
		if short != "heads/main" && short != "main" {
			t.Fatalf("ShortenRef = %q, want %q or %q", short, "heads/main", "main")
		}
	})
}

func TestResolveRefErrorSurface(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		repoHarness := testgit.NewRepo(t, testgit.RepoOptions{
			ObjectFormat: algo,
			Bare:         true,
			RefFormat:    "files",
		})

		root, err := os.OpenRoot(repoHarness.Dir())
		if err != nil {
			t.Fatalf("os.OpenRoot: %v", err)
		}
		defer func() { _ = root.Close() }()

		repo, err := repository.Open(root)
		if err != nil {
			t.Fatalf("repository.Open: %v", err)
		}
		defer func() { _ = repo.Close() }()

		_, err = repo.ResolveRef("refs/heads/does-not-exist")
		if err == nil {
			t.Fatalf("ResolveRef missing: expected error")
		}
		if !strings.Contains(err.Error(), "not found") {
			t.Fatalf("ResolveRef missing error = %v, want not found detail", err)
		}
	})
}

func TestListRefsLooseOverridesPacked(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		repoHarness := testgit.NewRepo(t, testgit.RepoOptions{
			ObjectFormat: algo,
			Bare:         true,
			RefFormat:    "files",
		})

		repoHarness.SymbolicRef(t, "HEAD", "refs/heads/main")
		_, _, commit1 := repoHarness.MakeCommit(t, "commit-one")
		repoHarness.UpdateRef(t, "refs/heads/main", commit1)
		repoHarness.UpdateRef(t, "refs/heads/feature", commit1)
		repoHarness.PackRefs(t, "--all", "--prune")

		_, _, commit2 := repoHarness.MakeCommit(t, "commit-two")
		repoHarness.UpdateRef(t, "refs/heads/main", commit2)

		root, err := os.OpenRoot(repoHarness.Dir())
		if err != nil {
			t.Fatalf("os.OpenRoot: %v", err)
		}
		defer func() { _ = root.Close() }()

		repo, err := repository.Open(root)
		if err != nil {
			t.Fatalf("repository.Open: %v", err)
		}
		defer func() { _ = repo.Close() }()

		mainRef, err := repo.ResolveRefFully("refs/heads/main")
		if err != nil {
			t.Fatalf("ResolveRefFully(main): %v", err)
		}
		if mainRef.ID != commit2 {
			t.Fatalf("ResolveRefFully(main) id = %s, want %s", mainRef.ID, commit2)
		}

		refs, err := repo.ListRefs("refs/heads/*")
		if err != nil {
			t.Fatalf("ListRefs(refs/heads/*): %v", err)
		}
		byName := make(map[string]ref.Ref, len(refs))
		for _, entry := range refs {
			name := entry.Name()
			if _, exists := byName[name]; exists {
				t.Fatalf("duplicate ref %q in ListRefs output", name)
			}
			byName[name] = entry
		}

		main, ok := byName["refs/heads/main"]
		if !ok {
			t.Fatalf("missing refs/heads/main in ListRefs output")
		}
		mainDetached, ok := main.(ref.Detached)
		if !ok {
			t.Fatalf("refs/heads/main type = %T, want ref.Detached", main)
		}
		if mainDetached.ID != commit2 {
			t.Fatalf("refs/heads/main id = %s, want %s", mainDetached.ID, commit2)
		}

		feature, ok := byName["refs/heads/feature"]
		if !ok {
			t.Fatalf("missing refs/heads/feature in ListRefs output")
		}
		featureDetached, ok := feature.(ref.Detached)
		if !ok {
			t.Fatalf("refs/heads/feature type = %T, want ref.Detached", feature)
		}
		if featureDetached.ID != commit1 {
			t.Fatalf("refs/heads/feature id = %s, want %s", featureDetached.ID, commit1)
		}
	})
}
