package testgit

import (
	"os"
	"testing"

	"codeberg.org/lindenii/furgit/objectid"
)

// RepoOptions controls git-init options for test repositories.
type RepoOptions struct {
	// Bare selects whether the repository is initialized as bare.
	Bare bool
	// RefFormat selects the git ref storage format (for example "files" or
	// "reftable"). Empty means git's default format.
	RefFormat string
}

// NewBareRepo creates a temporary bare repository initialized with the requested algorithm.
func NewBareRepo(tb testing.TB, algo objectid.Algorithm) *TestRepo {
	tb.Helper()
	return NewRepo(tb, algo, RepoOptions{Bare: true})
}

// NewWorkRepo creates a temporary non-bare repository initialized with the requested algorithm.
func NewWorkRepo(tb testing.TB, algo objectid.Algorithm) *TestRepo {
	tb.Helper()
	return NewRepo(tb, algo, RepoOptions{Bare: false})
}

// NewRepo creates a temporary repository initialized with the requested options.
func NewRepo(tb testing.TB, algo objectid.Algorithm, opts RepoOptions) *TestRepo {
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
	if opts.Bare {
		args = append(args, "--bare")
	}
	if opts.RefFormat != "" {
		args = append(args, "--ref-format="+opts.RefFormat)
	}
	args = append(args, dir)
	testRepo.runBytes(tb, nil, "", args...)
	return testRepo
}
