package repository_test

import (
	"errors"
	"io/fs"
	"os"
	"testing"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/repository"
)

func TestOpenClose(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo := newRepo(t, objectFormat)
			gitDir := openGitDir(t, repo)

			opened, err := repository.Open(gitDir, gitDir, repository.Options{})
			if err != nil {
				t.Fatalf("Open: %v", err)
			}

			err = opened.Close()
			if err != nil {
				t.Fatalf("Close: %v", err)
			}
		})
	}
}

func TestOpenMissingConfig(t *testing.T) {
	t.Parallel()

	root, err := os.OpenRoot(t.TempDir())
	if err != nil {
		t.Fatalf("OpenRoot: %v", err)
	}

	defer func() { _ = root.Close() }()

	_, err = repository.Open(root, root, repository.Options{})
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Open = %v, want fs.ErrNotExist", err)
	}
}
