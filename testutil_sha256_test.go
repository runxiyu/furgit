//go:build !sha1

package furgit

import (
	"os"
	"os/exec"
	"testing"
)

func setupTestRepo(t *testing.T) (string, func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "furgit-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	cleanup := func() {
		_ = os.RemoveAll(tempDir)
	}

	cmd := exec.Command("git", "init", "--object-format=sha256", "--bare", tempDir)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
	if output, err := cmd.CombinedOutput(); err != nil {
		cleanup()
		t.Fatalf("failed to init git repo: %v\n%s", err, output)
	}

	return tempDir, cleanup
}
