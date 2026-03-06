package testgit

import (
	"slices"
	"strings"
	"testing"

	"codeberg.org/lindenii/furgit/objectid"
)

// CommitTreeWithEnv creates one commit from a tree and message, optionally with
// parents, using additional environment variables for the git subprocess.
func (testRepo *TestRepo) CommitTreeWithEnv(
	tb testing.TB,
	extraEnv []string,
	tree objectid.ObjectID,
	message string,
	parents ...objectid.ObjectID,
) objectid.ObjectID {
	tb.Helper()

	args := make([]string, 0, 2+2*len(parents)+2)

	args = append(args, "commit-tree", tree.String())
	for _, parent := range parents {
		args = append(args, "-p", parent.String())
	}

	args = append(args, "-m", message)
	hex := testRepo.runWithExtraEnv(tb, extraEnv, args...)

	id, err := objectid.ParseHex(testRepo.algo, hex)
	if err != nil {
		tb.Fatalf("parse commit-tree output %q: %v", hex, err)
	}

	return id
}

func (testRepo *TestRepo) runWithExtraEnv(tb testing.TB, extraEnv []string, args ...string) string {
	tb.Helper()

	env := slices.Concat(testRepo.env, extraEnv)

	out, err := testRepo.runBytesWithEnv(tb, nil, testRepo.dir, env, args...)
	if err != nil {
		tb.Fatalf("git %v failed: %v\n%s", args, err, out)
	}

	return strings.TrimSpace(string(out))
}
