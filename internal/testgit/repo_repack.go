package testgit

import "testing"

// Repack runs "git repack" with args in the repository.
func (testRepo *TestRepo) Repack(tb testing.TB, args ...string) {
	tb.Helper()
	cmdArgs := make([]string, 0, len(args)+1)
	cmdArgs = append(cmdArgs, "repack")
	cmdArgs = append(cmdArgs, args...)
	_ = testRepo.Run(tb, cmdArgs...)
}
