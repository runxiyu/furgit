package testgit

import "testing"

// CommitGraphWrite runs "git commit-graph write" with args in the repository.
func (testRepo *TestRepo) CommitGraphWrite(tb testing.TB, args ...string) {
	tb.Helper()

	cmdArgs := make([]string, 0, len(args)+3)
	cmdArgs = append(cmdArgs, "commit-graph", "write")
	cmdArgs = append(cmdArgs, args...)
	_ = testRepo.Run(tb, cmdArgs...)
}
