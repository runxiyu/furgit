package testgit

import (
	"io"
	"testing"

	"codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/object/typ"
)

// HashObject hashes and writes an object and returns its object ID.
func (repo *Repo) HashObject(tb testing.TB, ty typ.Type, body io.Reader) id.ObjectID {
	tb.Helper()
	cmd := repo.Command(tb, "git", "hash-object", "-t", ty.Name(), "-w", "--stdin", "--literally")

	hex, err := cmd.CombinedOutput()
	if err != nil {
		tb.Fatalf("hash-object: %v", hex)
	}

	id, err := id.FromHex(repo.algo, string(hex))
	if err != nil {
		tb.Fatalf("parse git hash-object output %q: %v", hex, err)
	}

	return id
}
