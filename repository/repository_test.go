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

			opened, err := repository.Open(gitDir, gitDir)
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

	_, err = repository.Open(root, root)
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("Open = %v, want fs.ErrNotExist", err)
	}
}

func TestConfig(t *testing.T) {
	t.Parallel()

	repo := newRepo(t, id.ObjectFormatSHA1)

	err := repo.ConfigSet(t, "furgit.marker", "present")
	if err != nil {
		t.Fatalf("ConfigSet: %v", err)
	}

	opened := openRepository(t, repo)

	value, err := opened.Config().Lookup("furgit", "", "marker").String()
	if err != nil {
		t.Fatalf("Lookup: %v", err)
	}

	if value != "present" {
		t.Fatalf("marker = %q, want %q", value, "present")
	}
}
