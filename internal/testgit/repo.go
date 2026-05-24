package testgit

import (
	"os"
	"os/exec"
	"testing"

	"codeberg.org/lindenii/furgit/object/id"
)

type Repo struct {
	path string
	algo id.Algorithm
	env  []string
}

func (repo *Repo) Algorithm() id.Algorithm {
	return repo.algo
}

func (repo *Repo) Command(
	tb testing.TB,
	command string,
	args ...string,
) *exec.Cmd {
	//nolint:noctx
	cmd := exec.Command(command, args...)
	cmd.Dir = repo.path
	cmd.Env = repo.env

	return cmd
}

type RepoOptions struct {
	ObjectFormat id.Algorithm
}

func NewRepo(tb testing.TB, opts RepoOptions) (*Repo, error) {
	tb.Helper()

	repo := &Repo{
		path: tb.TempDir(),
		algo: opts.ObjectFormat,
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

	return repo, repo.Command(tb, "git", "init", "--object-format="+repo.algo.String(), "--", repo.path).Run()
}
