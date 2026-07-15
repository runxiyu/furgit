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
			opened := openRepository(t, repo, repository.Options{})

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
	openRepository(t, repo, repository.Options{})
}

func requireErrConfig(t *testing.T, objectFormat id.ObjectFormat, text string) {
	t.Helper()

	repo := newRepo(t, objectFormat)

	appendConfig(t, repo, text)

	gitDir := openGitDir(t, repo)

	_, err := repository.Open(gitDir, gitDir, repository.Options{})
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

func TestDuplicateRefStorageTakesLast(t *testing.T) {
	t.Parallel()

	requireOpens(
		t, id.ObjectFormatSHA256,
		"[extensions]\n\trefstorage = reftable\n\trefstorage = files\n",
	)
}

func TestDuplicateRefStorageTakesLastReftable(t *testing.T) {
	t.Parallel()

	requireErrConfig(
		t, id.ObjectFormatSHA256,
		"[extensions]\n\trefstorage = files\n\trefstorage = reftable\n",
	)
}

func TestRejectsRefStorageURI(t *testing.T) {
	t.Parallel()

	requireErrConfig(
		t, id.ObjectFormatSHA256,
		"[extensions]\n\trefstorage = files:///srv/references\n",
	)
}

func TestRejectsInvalidBooleanExtensionValues(t *testing.T) {
	t.Parallel()

	for _, name := range []string{
		"preciousobjects",
		"worktreeconfig",
		"relativeworktrees",
		"submodulepathconfig",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			requireErrConfig(
				t, id.ObjectFormatSHA256,
				"[extensions]\n\t"+name+" = nonsense\n",
			)
		})
	}
}

func TestExtensionValidationUsesEffectiveValue(t *testing.T) {
	t.Parallel()

	requireOpens(
		t, id.ObjectFormatSHA256,
		"[extensions]\n\trelativeworktrees = nonsense\n\trelativeworktrees = true\n",
	)
}

func TestRejectsValuelessPartialClone(t *testing.T) {
	t.Parallel()

	requireErrConfig(t, id.ObjectFormatSHA256, "[extensions]\n\tpartialclone\n")
}

func TestDuplicateObjectFormatTakesLast(t *testing.T) {
	t.Parallel()

	repo := newRepo(t, id.ObjectFormatSHA256)

	appendConfig(t, repo, "[extensions]\n\tobjectformat = sha1\n\tobjectformat = sha256\n")

	opened := openRepository(t, repo, repository.Options{})

	if opened.ObjectFormat() != id.ObjectFormatSHA256 {
		t.Fatalf("ObjectFormat = %v, want %v", opened.ObjectFormat(), id.ObjectFormatSHA256)
	}
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

	_, err = repository.Open(gitDir, gitDir, repository.Options{})
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

	_, err = repository.Open(gitDir, gitDir, repository.Options{})
	if !errors.Is(err, repository.ErrConfig) {
		t.Fatalf("Open = %v, want ErrConfig", err)
	}
}
