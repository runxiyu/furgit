package testgit

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

// Run executes git and returns trimmed textual output.
func (testRepo *TestRepo) Run(tb testing.TB, args ...string) string {
	tb.Helper()
	out := testRepo.runBytes(tb, nil, testRepo.dir, args...)

	return strings.TrimSpace(string(out))
}

// RunBytes executes git and returns raw output bytes.
func (testRepo *TestRepo) RunBytes(tb testing.TB, args ...string) []byte {
	tb.Helper()

	return testRepo.runBytes(tb, nil, testRepo.dir, args...)
}

// RunInput executes git with stdin and returns trimmed textual output.
func (testRepo *TestRepo) RunInput(tb testing.TB, stdin []byte, args ...string) string {
	tb.Helper()
	out := testRepo.runBytes(tb, stdin, testRepo.dir, args...)

	return strings.TrimSpace(string(out))
}

// RunInputBytes executes git with stdin and returns raw output bytes.
func (testRepo *TestRepo) RunInputBytes(tb testing.TB, stdin []byte, args ...string) []byte {
	tb.Helper()

	return testRepo.runBytes(tb, stdin, testRepo.dir, args...)
}

func (testRepo *TestRepo) runBytes(tb testing.TB, stdin []byte, dir string, args ...string) []byte {
	tb.Helper()
	//nolint:noctx
	cmd := exec.Command("git", args...) //#nosec G204
	cmd.Dir = dir

	cmd.Env = testRepo.env
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		tb.Fatalf("git %v failed: %v\n%s", args, err, out)
	}

	return out
}
