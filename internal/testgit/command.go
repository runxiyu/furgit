package testgit

import (
	"io"
	"os/exec"
	"slices"
	"strings"
	"testing"
)

func (repo *Repo) command(
	tb testing.TB,
	command string,
	args ...string,
) *exec.Cmd {
	tb.Helper()

	cmd := exec.CommandContext(tb.Context(), command, args...) //nolint:gosec
	cmd.Dir = repo.path
	cmd.Env = repo.env

	return cmd
}

func (repo *Repo) run(
	tb testing.TB,
	stdin io.Reader,
	command string, //nolint:unparam
	args ...string,
) (stdout []byte, err error) {
	tb.Helper()

	cmd := repo.command(tb, command, args...)

	cmd.Stdin = stdin

	return cmd.Output() //nolint:wrapcheck
}

func setEnv(env []string, key string, value string) []string {
	prefix := key + "="

	i := slices.IndexFunc(env, func(entry string) bool {
		return strings.HasPrefix(entry, prefix)
	})

	if i >= 0 {
		env[i] = prefix + value

		return env
	}

	return append(env, prefix+value)
}
