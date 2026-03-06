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

// RunE executes git and returns trimmed textual output plus any command error.
func (testRepo *TestRepo) RunE(tb testing.TB, args ...string) (string, error) {
	tb.Helper()

	out, err := testRepo.runBytesE(nil, testRepo.dir, args...)

	return strings.TrimSpace(string(out)), err
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

	out, err := testRepo.runBytesE(stdin, dir, args...)
	if err != nil {
		tb.Fatalf("git %v failed: %v\n%s", args, err, out)
	}

	return out
}

func (testRepo *TestRepo) runBytesE(stdin []byte, dir string, args ...string) ([]byte, error) {
	return testRepo.runBytesWithEnvNoHelper(stdin, dir, testRepo.env, args...)
}

// runBytesWithEnv executes git using the supplied environment.
func (testRepo *TestRepo) runBytesWithEnv(
	tb testing.TB,
	stdin []byte,
	dir string,
	env []string,
	args ...string,
) ([]byte, error) {
	tb.Helper()

	return testRepo.runBytesWithEnvNoHelper(stdin, dir, env, args...)
}

// runBytesWithEnvNoHelper executes git using the supplied environment without
// touching testing helper state.
func (testRepo *TestRepo) runBytesWithEnvNoHelper(
	stdin []byte,
	dir string,
	env []string,
	args ...string,
) ([]byte, error) {
	//nolint:noctx
	cmd := exec.Command("git", args...) //#nosec G204
	cmd.Dir = dir

	cmd.Env = env
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}

	return cmd.CombinedOutput()
}
