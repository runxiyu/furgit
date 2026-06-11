package testgit

import (
	"testing"
)

// VerifyPack runs git verify-pack -v on one pack index path
// and returns its raw verbose output.
func (repo *Repo) VerifyPack(tb testing.TB, idxPath string) ([]byte, error) {
	tb.Helper()

	return repo.run(tb, nil, "git", "verify-pack", "-v", "--end-of-options", idxPath)
}
