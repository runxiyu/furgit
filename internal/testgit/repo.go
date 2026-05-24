package testgit

import (
	"os"
	"os/exec"
	"testing"

	"codeberg.org/lindenii/furgit/object/id"
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

	repo := &Repo{
		path:         tb.TempDir(),
		objectFormat: opts.ObjectFormat,
		env: append(os.Environ(),
			"GIT_CONFIG_GLOBAL=/dev/null",
			"GIT_CONFIG_SYSTEM=/dev/null",
			"GIT_AUTHOR_NAME=Test Author",
			"GIT_AUTHOR_EMAIL=test@example.org",
			"GIT_COMMITTER_NAME=Test Committer",
			"GIT_COMMITTER_EMAIL=committer@example.org",
			"GIT_AUTHOR_DATE=1234567890 +0000",
			"GIT_COMMITTER_DATE=1234567890 +0000",
		),
	}

	return repo, repo.Command(tb, "git", "init", "--object-format="+repo.objectFormat.String(), "--", repo.path).Run() //nolint:wrapcheck
}

func (repo *Repo) ObjectFormat() id.ObjectFormat {
	return repo.objectFormat
}

func (repo *Repo) Command(
	tb testing.TB,
	command string,
	args ...string,
) *exec.Cmd {
	tb.Helper()

	//nolint:noctx
	cmd := exec.Command(command, args...)
	cmd.Dir = repo.path
	cmd.Env = repo.env

	return cmd
}
