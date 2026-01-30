package furgit

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func setupWorkDir(t *testing.T) (string, func()) {
	t.Helper()
	workDir, err := os.MkdirTemp("", "furgit-work-*")
	if err != nil {
		t.Fatalf("failed to create work dir: %v", err)
	}
	return workDir, func() { _ = os.RemoveAll(workDir) }
}

func gitCmd(t *testing.T, dir string, stdin []byte, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_CONFIG_SYSTEM=/dev/null",
		"GIT_AUTHOR_NAME=Test Author",
		"GIT_AUTHOR_EMAIL=test@example.org",
		"GIT_COMMITTER_NAME=Test Committer",
		"GIT_COMMITTER_EMAIL=committer@example.org",
		"GIT_AUTHOR_DATE=1234567890 +0000",
		"GIT_COMMITTER_DATE=1234567890 +0000",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, output)
	}
	return strings.TrimSpace(string(output))
}

func gitHashObject(t *testing.T, dir, objType string, data []byte) string {
	t.Helper()
	cmd := exec.Command("git", "hash-object", "-t", objType, "-w", "--stdin")
	cmd.Dir = dir
	cmd.Stdin = bytes.NewReader(data)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git hash-object failed: %v\n%s", err, output)
	}
	return strings.TrimSpace(string(output))
}

func gitCatFile(t *testing.T, dir, objType, hash string) []byte {
	t.Helper()
	cmd := exec.Command("git", "cat-file", objType, hash)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git cat-file %s %s failed: %v\n%s", objType, hash, err, output)
	}
	if objType == "-t" || objType == "-s" {
		return bytes.TrimSpace(output)
	}
	return output
}
