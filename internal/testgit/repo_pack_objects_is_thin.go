package testgit

import (
	"os/exec"
	"strings"
	"testing"
)

// PackObjectsIsThin reports whether git emits one thin pack for the given revs.
//
// It streams `git pack-objects --stdout --revs --thin` into `git index-pack
// --stdin` in one scratch bare repository. A failure in index-pack due to
// unresolved deltas is treated as confirmation that the emitted pack is thin.
func (testRepo *TestRepo) PackObjectsIsThin(tb testing.TB, revs []string) bool {
	tb.Helper()

	scratch := NewRepo(tb, RepoOptions{ObjectFormat: testRepo.algo, Bare: true})

	packArgs := []string{"pack-objects", "--stdout", "--revs", "--thin"}
	//nolint:noctx
	packCmd := exec.Command("git", packArgs...) //#nosec G204
	packCmd.Dir = testRepo.dir
	packCmd.Env = testRepo.env
	packCmd.Stdin = strings.NewReader(strings.Join(revs, "\n") + "\n")
	packStderr := &strings.Builder{}
	packCmd.Stderr = packStderr

	packStdout, err := packCmd.StdoutPipe()
	if err != nil {
		tb.Fatalf("git %v stdout pipe: %v", packArgs, err)
	}

	indexArgs := []string{"index-pack", "--stdin"}
	//nolint:noctx
	indexCmd := exec.Command("git", indexArgs...) //#nosec G204
	indexCmd.Dir = scratch.dir
	indexCmd.Env = scratch.env
	indexCmd.Stdin = packStdout
	indexStderr := &strings.Builder{}
	indexCmd.Stderr = indexStderr

	err = indexCmd.Start()
	if err != nil {
		tb.Fatalf("git %v start failed: %v", indexArgs, err)
	}

	err = packCmd.Start()
	if err != nil {
		_ = indexCmd.Process.Kill()
		_ = indexCmd.Wait()

		tb.Fatalf("git %v start failed: %v", packArgs, err)
	}

	packErr := packCmd.Wait()
	if packErr != nil {
		tb.Fatalf("git %v failed: %v\n%s", packArgs, packErr, packStderr.String())
	}

	indexErr := indexCmd.Wait()
	if indexErr == nil {
		return false
	}

	stderr := strings.ToLower(indexStderr.String())
	if strings.Contains(stderr, "unresolved") && strings.Contains(stderr, "delta") {
		return true
	}

	if strings.Contains(stderr, "missing") && strings.Contains(stderr, "base") {
		return true
	}

	tb.Fatalf("git %v failed unexpectedly: %v\n%s", indexArgs, indexErr, indexStderr.String())

	return false
}
