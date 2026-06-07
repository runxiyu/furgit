package testgit

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/typ"
)

// HashObject hashes and writes an object,
// and returns its object ID.
func (repo *Repo) HashObject(tb testing.TB, ty typ.Type, body io.Reader) (id.ObjectID, error) {
	tb.Helper()

	stdout, err := repo.run(tb, body, "git", "hash-object", "-t", ty.Name(), "-w", "--stdin", "--literally")
	if err != nil {
		return id.ObjectID{}, fmt.Errorf("hash-object: %w", err)
	}

	objectID, err := repo.objectFormat.FromString(strings.TrimSuffix(string(stdout), "\n"))
	if err != nil {
		return id.ObjectID{}, fmt.Errorf("parse git hash-object output %q: %w", string(stdout), err)
	}

	return objectID, nil
}

// CatFile returns the raw content of an object
// (without the "type size\x00" header).
func (repo *Repo) CatFile(tb testing.TB, ty typ.Type, oid id.ObjectID) ([]byte, error) {
	tb.Helper()

	stdout, err := repo.run(tb, nil, "git", "cat-file", ty.Name(), "--end-of-options", oid.String())
	if err != nil {
		return nil, fmt.Errorf("cat-file: %w", err)
	}

	return stdout, nil
}
