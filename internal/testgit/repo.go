package testgit

import (
	"os"
	"testing"

	"lindenii.org/go/furgit/object/id"
)

type Repo struct {
	path         string
	objectFormat id.ObjectFormat
	env          []string
}

type RepoOptions struct {
	ObjectFormat id.ObjectFormat
}

func NewRepo(tb testing.TB, opts RepoOptions) (*Repo, error) {
	tb.Helper()

	objectFormat := opts.ObjectFormat
	if objectFormat == id.ObjectFormatUnknown {
		objectFormat = id.ObjectFormatSHA256
	}

	repo := &Repo{
		path:         tb.TempDir(),
		objectFormat: objectFormat,
		env: append(os.Environ(),
			"GIT_CONFIG_GLOBAL=/dev/null",
			"GIT_CONFIG_SYSTEM=/dev/null",
			"GIT_AUTHOR_NAME=Test Author",
			"GIT_AUTHOR_EMAIL=author@example.org",
			"GIT_COMMITTER_NAME=Test Committer",
			"GIT_COMMITTER_EMAIL=committer@example.org",
			"GIT_AUTHOR_DATE=1234567890 +0000",
			"GIT_COMMITTER_DATE=1234567890 +0000",
		),
	}

	return repo, repo.Command(tb, "git", "init", "--object-format="+repo.objectFormat.String(), "--end-of-options", repo.path).Run() //nolint:wrapcheck
}

func (repo *Repo) ObjectFormat(tb testing.TB) id.ObjectFormat {
	tb.Helper()

	return repo.objectFormat
}

func (repo *Repo) Root(tb testing.TB) *os.Root {
	tb.Helper()

	root, err := os.OpenRoot(repo.path)
	if err != nil {
		tb.Fatalf("failed opening repo root at %q: %v", repo.path, err)
	}

	return root
}
