package testgit

import (
	"io"
	"testing"

	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/typ"
)

// HashObject hashes and writes an object,
// and returns its object ID.
func (repo *Repo) HashObject(tb testing.TB, ty typ.Type, body io.Reader) id.ObjectID {
	tb.Helper()

	stdout, err := repo.Run(tb, body, "git", "hash-object", "-t", ty.Name(), "-w", "--stdin", "--literally")
	if err != nil {
		tb.Fatalf("hash-object: %v", err)
	}

	id, err := repo.objectFormat.FromString(string(stdout))
	if err != nil {
		tb.Fatalf("parse git hash-object output %q: %v", string(stdout), err)
	}

	return id
}
