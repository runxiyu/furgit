package repository_test

import (
	"os"
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/repository"
)

func newRepo(t *testing.T, objectFormat id.ObjectFormat) *testgit.Repo {
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

func openGitDir(t *testing.T, repo *testgit.Repo) *os.Root {
	t.Helper()

	root := repo.Root(t)

	t.Cleanup(func() { _ = root.Close() })

	gitDir, err := root.OpenRoot(".git")
	if err != nil {
		t.Fatalf("OpenRoot(.git): %v", err)
	}

	t.Cleanup(func() { _ = gitDir.Close() })

	return gitDir
}

func openRepository(t *testing.T, repo *testgit.Repo) *repository.Repository {
	t.Helper()

	gitDir := openGitDir(t, repo)

	opened, err := repository.Open(gitDir, gitDir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	t.Cleanup(func() { _ = opened.Close() })

	return opened
}

func appendConfig(t *testing.T, repo *testgit.Repo, text string) {
	t.Helper()

	root := repo.Root(t)

	t.Cleanup(func() { _ = root.Close() })

	file, err := root.OpenFile(".git/config", os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		t.Fatalf("open config: %v", err)
	}

	defer func() { _ = file.Close() }()

	_, err = file.WriteString(text)
	if err != nil {
		t.Fatalf("write config: %v", err)
	}
}

func makeCommit(t *testing.T, repo *testgit.Repo, message string) id.ObjectID {
	t.Helper()

	tree, err := repo.MkTree(t, nil) // again, we don't care about contents
	if err != nil {
		t.Fatalf("MkTree: %v", err)
	}

	commit, err := repo.CommitTree(t, tree, testgit.CommitTreeOptions{Message: message})
	if err != nil {
		t.Fatalf("CommitTree: %v", err)
	}

	return commit
}
