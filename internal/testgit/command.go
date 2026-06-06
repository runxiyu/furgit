package testgit

import (
	"io"
	"os/exec"
	"testing"
)

func (repo *Repo) Command(
	tb testing.TB,
	command string,
	args ...string,
) *exec.Cmd {
	tb.Helper()

	cmd := exec.CommandContext(tb.Context(), command, args...)
	cmd.Dir = repo.path
	cmd.Env = repo.env

	return cmd
}

func (repo *Repo) Run(
	tb testing.TB,
	stdin io.Reader,
	command string,
	args ...string,
) (stdout []byte, err error) {
	tb.Helper()

	cmd := repo.Command(tb, command, args...)

	cmd.Stdin = stdin

	return cmd.Output() //nolint:wrapcheck
}
