package repository_test

import (
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

		repo, err := repository.Open(repoHarness.Dir())
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

		repo, err := repository.Open(repoHarness.Dir())
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
