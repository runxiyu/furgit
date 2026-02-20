package testgit

import (
	"testing"

	"codeberg.org/lindenii/furgit/oid"
)

// HashObject hashes and writes an object and returns its object ID.
func (repo *TestRepo) HashObject(tb testing.TB, objType string, body []byte) oid.ObjectID {
	tb.Helper()
	hex := repo.RunInput(tb, body, "hash-object", "-t", objType, "-w", "--stdin")
	id, err := oid.ParseHex(repo.algo, hex)
	if err != nil {
		tb.Fatalf("parse git hash-object output %q: %v", hex, err)
	}
	return id
}
