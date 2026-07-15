package repository_test

import (
	"testing"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/typ"
	"lindenii.org/go/furgit/repository"
)

func TestObjects(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo := newRepo(t, objectFormat)
			commitID := makeCommit(t, repo, "objects")

			opened := openRepository(t, repo, repository.Options{})

			objectType, _, err := opened.Objects().ReadBytesContent(commitID)
			if err != nil {
				t.Fatalf("ReadBytesContent: %v", err)
			}

			if objectType != typ.Commit {
				t.Fatalf("type = %v, want %v", objectType, typ.Commit)
			}
		})
	}
}

func TestFetcher(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo := newRepo(t, objectFormat)
			commitID := makeCommit(t, repo, "fetcher")

			opened := openRepository(t, repo, repository.Options{})

			fetched, err := opened.Fetcher().ExactCommit(commitID)
			if err != nil {
				t.Fatalf("ExactCommit: %v", err)
			}

			if fetched.ID() != commitID {
				t.Fatalf("ID = %v, want %v", fetched.ID(), commitID)
			}

			message := string(fetched.Object().Message)
			if message != "fetcher\n" {
				t.Fatalf("Message = %q, want %q", message, "fetcher\n")
			}
		})
	}
}

func TestOpenCreatesPackDirectory(t *testing.T) {
	t.Parallel()

	repo := newRepo(t, id.ObjectFormatSHA256)
	root := repo.Root(t)

	t.Cleanup(func() { _ = root.Close() })

	err := root.RemoveAll(".git/objects/pack")
	if err != nil {
		t.Fatalf("RemoveAll: %v", err)
	}

	gitDir := openGitDir(t, repo)

	opened, err := repository.Open(gitDir, gitDir, repository.Options{})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	t.Cleanup(func() { _ = opened.Close() })

	_, err = root.Stat(".git/objects/pack")
	if err != nil {
		t.Fatalf("Stat objects/pack: %v", err)
	}
}
