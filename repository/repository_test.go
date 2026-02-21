package repository_test

import (
	"testing"

	"codeberg.org/lindenii/furgit/internal/testgit"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
	"codeberg.org/lindenii/furgit/ref"
	"codeberg.org/lindenii/furgit/repository"
)

func TestOpenFilesRefFormat(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		repoHarness := testgit.NewRepo(t, testgit.RepoOptions{
			ObjectFormat: algo,
			Bare:         true,
			RefFormat:    "files",
		})

		_, _, commitID := repoHarness.MakeCommit(t, "files refs")
		repoHarness.UpdateRef(t, "refs/heads/main", commitID)
		repoHarness.SymbolicRef(t, "HEAD", "refs/heads/main")

		repo, err := repository.Open(repoHarness.Dir())
		if err != nil {
			t.Fatalf("repository.Open: %v", err)
		}
		defer func() { _ = repo.Close() }()

		if repo.Algorithm() != algo {
			t.Fatalf("Algorithm = %v, want %v", repo.Algorithm(), algo)
		}

		headerType, headerSize, err := repo.Objects().ReadHeader(commitID)
		if err != nil {
			t.Fatalf("ReadHeader(commit): %v", err)
		}
		if headerType != objecttype.TypeCommit {
			t.Fatalf("ReadHeader(commit) type = %v, want %v", headerType, objecttype.TypeCommit)
		}
		if headerSize <= 0 {
			t.Fatalf("ReadHeader(commit) size = %d, want > 0", headerSize)
		}

		resolved, err := repo.Refs().Resolve("refs/heads/main")
		if err != nil {
			t.Fatalf("Resolve(refs/heads/main): %v", err)
		}
		detached, ok := resolved.(ref.Detached)
		if !ok {
			t.Fatalf("Resolve(refs/heads/main) type = %T, want ref.Detached", resolved)
		}
		if detached.ID != commitID {
			t.Fatalf("Resolve(refs/heads/main) id = %s, want %s", detached.ID, commitID)
		}

		head, err := repo.Refs().ResolveFully("HEAD")
		if err != nil {
			t.Fatalf("ResolveFully(HEAD): %v", err)
		}
		if head.ID != commitID {
			t.Fatalf("ResolveFully(HEAD) id = %s, want %s", head.ID, commitID)
		}
	})
}

func TestOpenFilesWithPackedRefs(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		repoHarness := newRepoForRefs(t, algo, "files")
		commitID := writeMainAndHead(t, repoHarness)
		repoHarness.PackRefs(t, "--all", "--prune")
		assertResolveFully(t, repoHarness, "refs/heads/main", commitID)
	})
}

func TestOpenReftableRefFormat(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		repoHarness := newRepoForRefs(t, algo, "reftable")
		commitID := writeMainAndHead(t, repoHarness)
		assertResolveFully(t, repoHarness, "HEAD", commitID)
	})
}

func newRepoForRefs(t *testing.T, algo objectid.Algorithm, refFormat string) *testgit.TestRepo {
	t.Helper()
	return testgit.NewRepo(t, testgit.RepoOptions{
		ObjectFormat: algo,
		Bare:         true,
		RefFormat:    refFormat,
	})
}

func writeMainAndHead(t *testing.T, repoHarness *testgit.TestRepo) objectid.ObjectID {
	t.Helper()
	_, _, commitID := repoHarness.MakeCommit(t, "refs")
	repoHarness.UpdateRef(t, "refs/heads/main", commitID)
	repoHarness.SymbolicRef(t, "HEAD", "refs/heads/main")
	return commitID
}

func assertResolveFully(t *testing.T, repoHarness *testgit.TestRepo, name string, want objectid.ObjectID) {
	t.Helper()

	repo, err := repository.Open(repoHarness.Dir())
	if err != nil {
		t.Fatalf("repository.Open: %v", err)
	}
	defer func() { _ = repo.Close() }()

	resolved, err := repo.Refs().ResolveFully(name)
	if err != nil {
		t.Fatalf("ResolveFully(%s): %v", name, err)
	}
	if resolved.ID != want {
		t.Fatalf("ResolveFully(%s) id = %s, want %s", name, resolved.ID, want)
	}
}
