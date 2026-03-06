package testgit

import (
	"io/fs"
	"os"
	"testing"

	"codeberg.org/lindenii/furgit/objectid"
)

// NewRepoFromFixture copies one existing repository fixture into a temp dir.
func NewRepoFromFixture(tb testing.TB, algo objectid.Algorithm, fixtureDir string) *TestRepo {
	tb.Helper()

	if algo.Size() == 0 {
		tb.Fatalf("invalid algorithm: %v", algo)
	}

	dst := tb.TempDir()
	srcFS := os.DirFS(fixtureDir)

	err := copyFS(dst, srcFS)
	if err != nil {
		tb.Fatalf("copy fixture %q: %v", fixtureDir, err)
	}

	return &TestRepo{
		dir:  dst,
		algo: algo,
		env:  defaultEnv(),
	}
}

func copyFS(dst string, src fs.FS) error {
	return os.CopyFS(dst, src)
}
