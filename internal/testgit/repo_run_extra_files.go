package testgit

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"testing"
)

// RunWithExtraFilesE executes git with inherited extra files and returns split
// stdout/stderr plus any command error.
func (testRepo *TestRepo) RunWithExtraFilesE(
	tb testing.TB,
	extraFiles []*os.File,
	args ...string,
) ([]byte, []byte, error) {
	tb.Helper()

	return testRepo.RunWithExtraFilesEnvContextE(
		tb,
		context.Background(),
		nil,
		extraFiles,
		args...,
	)
}

// RunWithExtraFilesEnvContextE executes git with inherited extra files, extra
// environment, and context cancellation, returning split stdout/stderr plus any
// command error.
func (testRepo *TestRepo) RunWithExtraFilesEnvContextE(
	tb testing.TB,
	ctx context.Context,
	extraEnv []string,
	extraFiles []*os.File,
	args ...string,
) ([]byte, []byte, error) {
	tb.Helper()

	cmd := exec.CommandContext(ctx, "git", args...) //#nosec G204
	cmd.Dir = testRepo.dir
	cmd.Env = testRepo.env
	cmd.Env = append(cmd.Env, extraEnv...)
	cmd.ExtraFiles = append([]*os.File(nil), extraFiles...)

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	return stdout.Bytes(), stderr.Bytes(), err
}
