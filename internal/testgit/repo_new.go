package testgit

import (
	"os"
	"testing"

	"codeberg.org/lindenii/furgit/objectid"
)

// NewBareRepo creates a temporary bare repository initialized with the requested algorithm.
func NewBareRepo(tb testing.TB, algo objectid.Algorithm) *TestRepo {
	tb.Helper()
	return newRepo(tb, algo, true)
}

// NewWorkRepo creates a temporary non-bare repository initialized with the requested algorithm.
func NewWorkRepo(tb testing.TB, algo objectid.Algorithm) *TestRepo {
	tb.Helper()
	return newRepo(tb, algo, false)
}

func newRepo(tb testing.TB, algo objectid.Algorithm, bare bool) *TestRepo {
	tb.Helper()
	if algo.Size() == 0 {
		tb.Fatalf("invalid algorithm: %v", algo)
	}

	dir, err := os.MkdirTemp("", "furgit-testgit-*")
	if err != nil {
		tb.Fatalf("create temp dir: %v", err)
	}
	tb.Cleanup(func() { _ = os.RemoveAll(dir) })

	testRepo := &TestRepo{
		dir:  dir,
		algo: algo,
		env: append(os.Environ(),
			"GIT_CONFIG_GLOBAL=/dev/null",
			"GIT_CONFIG_SYSTEM=/dev/null",
			"GIT_AUTHOR_NAME=Test Author",
			"GIT_AUTHOR_EMAIL=test@example.org",
			"GIT_COMMITTER_NAME=Test Committer",
			"GIT_COMMITTER_EMAIL=committer@example.org",
			"GIT_AUTHOR_DATE=1234567890 +0000",
			"GIT_COMMITTER_DATE=1234567890 +0000",
		),
	}

	args := []string{"init", "--object-format=" + algo.String()}
	if bare {
		args = append(args, "--bare")
	}
	args = append(args, dir)
	testRepo.runBytes(tb, nil, "", args...)
	return testRepo
}
