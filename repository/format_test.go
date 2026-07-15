package repository_test

import (
	"errors"
	"testing"

	"lindenii.org/go/furgit/internal/testgit"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/repository"
)

func TestObjectFormat(t *testing.T) {
	t.Parallel()

	for _, objectFormat := range id.SupportedObjectFormats() {
		t.Run(objectFormat.String(), func(t *testing.T) {
			t.Parallel()

			repo := newRepo(t, objectFormat)
			opened := openRepository(t, repo)

			if opened.ObjectFormat() != objectFormat {
				t.Fatalf("ObjectFormat = %v, want %v", opened.ObjectFormat(), objectFormat)
			}
		})
	}
}

func requireOpens(t *testing.T, objectFormat id.ObjectFormat, text string) {
	t.Helper()

	repo := newRepo(t, objectFormat)

	appendConfig(t, repo, text)
	openRepository(t, repo)
}

func requireErrConfig(t *testing.T, objectFormat id.ObjectFormat, text string) {
	t.Helper()

	repo := newRepo(t, objectFormat)

	appendConfig(t, repo, text)

	gitDir := openGitDir(t, repo)

	_, err := repository.Open(gitDir, gitDir)
	if !errors.Is(err, repository.ErrConfig) {
		t.Fatalf("Open = %v, want ErrConfig", err)
	}
}

func TestVersion0IgnoresUnknownExtension(t *testing.T) {
	t.Parallel()

	requireOpens(t, id.ObjectFormatSHA1, "[extensions]\n\tfurgitnonsense = yes\n")
}

func TestVersion1RejectsUnknownExtension(t *testing.T) {
	t.Parallel()

	requireErrConfig(t, id.ObjectFormatSHA256, "[extensions]\n\tfurgitnonsense = yes\n")
}

func TestVersion0RejectsVersion1Extension(t *testing.T) {
	t.Parallel()

	requireErrConfig(t, id.ObjectFormatSHA1, "[extensions]\n\trelativeworktrees = true\n")
}

func TestAnyVersionExtensionsAccepted(t *testing.T) {
	t.Parallel()

	requireOpens(
		t, id.ObjectFormatSHA256,
		"[extensions]\n\tpreciousobjects = true\n\tworktreeconfig = true\n",
	)
}

func TestRejectsCompatObjectFormat(t *testing.T) {
	t.Parallel()

	requireErrConfig(t, id.ObjectFormatSHA256, "[extensions]\n\tcompatobjectformat = sha1\n")
}

func TestRejectsInvalidObjectFormat(t *testing.T) {
	t.Parallel()

	requireErrConfig(t, id.ObjectFormatSHA256, "[extensions]\n\tobjectformat = susnsa\n")
}

func TestRejectsFutureVersion(t *testing.T) {
	t.Parallel()

	repo := newRepo(t, id.ObjectFormatSHA1)

	err := repo.ConfigSet(t, "core.repositoryformatversion", "2")
	if err != nil {
		t.Fatalf("ConfigSet: %v", err)
	}

	gitDir := openGitDir(t, repo)

	_, err = repository.Open(gitDir, gitDir)
	if !errors.Is(err, repository.ErrConfig) {
		t.Fatalf("Open = %v, want ErrConfig", err)
	}
}

func TestRejectsReftable(t *testing.T) {
	// TODO

	t.Parallel()

	repo, err := testgit.NewRepo(t, testgit.RepoOptions{
		ObjectFormat: id.ObjectFormatSHA256,
		RefFormat:    "reftable",
	})
	if err != nil {
		t.Fatalf("NewRepo: %v", err)
	}

	gitDir := openGitDir(t, repo)

	_, err = repository.Open(gitDir, gitDir)
	if !errors.Is(err, repository.ErrConfig) {
		t.Fatalf("Open = %v, want ErrConfig", err)
	}
}
