package files_test

import (
	"os"
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/ref/store/files"
)

func newFilesRepo(t *testing.T, objectFormat id.ObjectFormat) *testgit.Repo {
	t.Helper()

	repo, err := testgit.NewRepo(t, testgit.RepoOptions{
		ObjectFormat: objectFormat,
		RefFormat:    "files",
	})
	if err != nil {
		t.Fatalf("NewRepo: %v", err)
	}

	return repo
}

func openGitDirRoot(t *testing.T, repo *testgit.Repo) *os.Root {
	t.Helper()

	root := repo.Root(t)

	t.Cleanup(func() { _ = root.Close() })

	gitdir, err := root.OpenRoot(".git")
	if err != nil {
		t.Fatalf("OpenRoot(.git): %v", err)
	}

	t.Cleanup(func() { _ = gitdir.Close() })

	return gitdir
}

func openStore(t *testing.T, repo *testgit.Repo, objectFormat id.ObjectFormat) *files.Files {
	t.Helper()

	gitdir := openGitDirRoot(t, repo)

	filesStore := files.New(gitdir, gitdir, objectFormat, files.Options{
		LooseLockTimeout:  files.DefaultLooseLockTimeout,
		PackedLockTimeout: files.DefaultPackedLockTimeout,
	})

	t.Cleanup(func() { _ = filesStore.Close() })

	return filesStore
}

func makeCommit(t *testing.T, repo *testgit.Repo, message string) id.ObjectID {
	t.Helper()

	tree, err := repo.MkTree(t, nil) // we don't care about contents
	if err != nil {
		t.Fatalf("MkTree: %v", err)
	}

	commit, err := repo.CommitTree(t, tree, testgit.CommitTreeOptions{Message: message})
	if err != nil {
		t.Fatalf("CommitTree: %v", err)
	}

	return commit
}
