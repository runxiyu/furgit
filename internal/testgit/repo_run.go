package testgit

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

// Run executes git and returns trimmed textual output.
func (repo *TestRepo) Run(tb testing.TB, args ...string) string {
	tb.Helper()
	out := repo.runBytes(tb, nil, repo.dir, args...)
	return strings.TrimSpace(string(out))
}

// RunBytes executes git and returns raw output bytes.
func (repo *TestRepo) RunBytes(tb testing.TB, args ...string) []byte {
	tb.Helper()
	return repo.runBytes(tb, nil, repo.dir, args...)
}

// RunInput executes git with stdin and returns trimmed textual output.
func (repo *TestRepo) RunInput(tb testing.TB, stdin []byte, args ...string) string {
	tb.Helper()
	out := repo.runBytes(tb, stdin, repo.dir, args...)
	return strings.TrimSpace(string(out))
}

// RunInputBytes executes git with stdin and returns raw output bytes.
func (repo *TestRepo) RunInputBytes(tb testing.TB, stdin []byte, args ...string) []byte {
	tb.Helper()
	return repo.runBytes(tb, stdin, repo.dir, args...)
}

func (repo *TestRepo) runBytes(tb testing.TB, stdin []byte, dir string, args ...string) []byte {
	tb.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = repo.env
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		tb.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
	return out
}
