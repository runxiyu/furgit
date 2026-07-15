package testgit

import (
	"fmt"
	"testing"

	"lindenii.org/go/furgit/object/id"
)

// UpdateRef points a ref at an object so that ref-aware tooling can see it.
func (repo *Repo) UpdateRef(tb testing.TB, name string, oid id.ObjectID) error {
	tb.Helper()

	_, err := repo.run(tb, nil, "git", "update-ref", "--end-of-options", name, oid.String())
	if err != nil {
		return fmt.Errorf("update-ref %s %s: %w", name, oid, err)
	}

	return nil
}

// SymbolicRefUpdate points a symbolic ref at a target reference name.
func (repo *Repo) SymbolicRefUpdate(tb testing.TB, name, target string) error {
	tb.Helper()

	_, err := repo.run(tb, nil, "git", "symbolic-ref", "--end-of-options", name, target)
	if err != nil {
		return fmt.Errorf("symbolic-ref %s %s: %w", name, target, err)
	}

	return nil
}

// PackRefs packs all refs into packed-refs and prunes the loose copies.
func (repo *Repo) PackRefs(tb testing.TB) error {
	tb.Helper()

	_, err := repo.run(tb, nil, "git", "pack-refs", "--all", "--prune")
	if err != nil {
		return fmt.Errorf("pack-refs: %w", err)
	}

	return nil
}

// ForEachRefFormat returns git-for-each-ref output for one ref and one format.
func (repo *Repo) ForEachRefFormat(tb testing.TB, pattern string, format string) ([]byte, error) {
	tb.Helper()

	stdout, err := repo.run(tb, nil, "git", "for-each-ref", "--format="+format, "--end-of-options", pattern)
	if err != nil {
		return nil, fmt.Errorf("for-each-ref --format=%q %s: %w", format, pattern, err)
	}

	return stdout, nil
}
