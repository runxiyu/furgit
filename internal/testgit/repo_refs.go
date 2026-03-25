package testgit

import (
	"strings"
	"testing"

	objectid "codeberg.org/lindenii/furgit/object/id"
)

// UpdateRef updates a ref to point at id.
func (testRepo *TestRepo) UpdateRef(tb testing.TB, name string, id objectid.ObjectID) {
	tb.Helper()
	testRepo.Run(tb, "update-ref", name, id.String())
}

// DeleteRef deletes a ref.
func (testRepo *TestRepo) DeleteRef(tb testing.TB, name string) {
	tb.Helper()
	testRepo.Run(tb, "update-ref", "-d", name)
}

// SymbolicRef sets a symbolic reference target.
func (testRepo *TestRepo) SymbolicRef(tb testing.TB, name, target string) {
	tb.Helper()
	testRepo.Run(tb, "symbolic-ref", name, target)
}

// PackRefs runs git pack-refs with args.
func (testRepo *TestRepo) PackRefs(tb testing.TB, args ...string) {
	tb.Helper()

	cmd := append([]string{"pack-refs"}, args...)
	testRepo.Run(tb, cmd...)
}

// ShowRef returns lines from git show-ref output.
func (testRepo *TestRepo) ShowRef(tb testing.TB, args ...string) []string {
	tb.Helper()

	cmd := append([]string{"show-ref"}, args...)

	out := testRepo.Run(tb, cmd...)
	if strings.TrimSpace(out) == "" {
		return nil
	}

	return strings.Split(strings.TrimSpace(out), "\n")
}
