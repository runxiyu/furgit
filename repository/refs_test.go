package repository_test

import (
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	objectid "lindenii.org/go/furgit/object/id"
	objecttype "lindenii.org/go/furgit/object/type"
	"lindenii.org/go/furgit/ref"
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

		repo := repoHarness.OpenRepository(t)

		if repo.Algorithm() != algo {
			t.Fatalf("Algorithm = %v, want %v", repo.Algorithm(), algo)
		}

		headerType, headerSize, err := repo.ObjectStore().ReadHeader(commitID)
		if err != nil {
			t.Fatalf("ReadHeader(commit): %v", err)
		}

		if headerType != objecttype.TypeCommit {
			t.Fatalf("ReadHeader(commit) type = %v, want %v", headerType, objecttype.TypeCommit)
		}

		if headerSize <= 0 {
			t.Fatalf("ReadHeader(commit) size = %d, want > 0", headerSize)
		}

		resolved, err := repo.RefStore().Resolve("refs/heads/main")
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

		head, err := repo.RefStore().ResolveToDetached("HEAD")
		if err != nil {
			t.Fatalf("ResolveToDetached(HEAD): %v", err)
		}

		if head.ID != commitID {
			t.Fatalf("ResolveToDetached(HEAD) id = %s, want %s", head.ID, commitID)
		}
	})
}

func TestOpenFilesWithPackedRefs(t *testing.T) {
	t.Parallel()

	testgit.ForEachAlgorithm(t, func(t *testing.T, algo objectid.Algorithm) { //nolint:thelper
		repoHarness := newRepoForRefs(t, algo, "files")
		commitID := writeMainAndHead(t, repoHarness)
		repoHarness.PackRefs(t, "--all", "--prune")
		assertResolveToDetached(t, repoHarness, "refs/heads/main", commitID)
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

func assertResolveToDetached(t *testing.T, repoHarness *testgit.TestRepo, name string, want objectid.ObjectID) {
	t.Helper()

	repo := repoHarness.OpenRepository(t)

	resolved, err := repo.RefStore().ResolveToDetached(name)
	if err != nil {
		t.Fatalf("ResolveToDetached(%s): %v", name, err)
	}

	if resolved.ID != want {
		t.Fatalf("ResolveToDetached(%s) id = %s, want %s", name, resolved.ID, want)
	}
}
