package testgit

import (
	"testing"

	objectid "lindenii.org/go/furgit/object/id"
)

// HashObject hashes and writes an object and returns its object ID.
func (testRepo *TestRepo) HashObject(tb testing.TB, objType string, body []byte) objectid.ObjectID {
	tb.Helper()
	hex := testRepo.RunInput(tb, body, "hash-object", "-t", objType, "-w", "--stdin")

	id, err := objectid.ParseHex(testRepo.algo, hex)
	if err != nil {
		tb.Fatalf("parse git hash-object output %q: %v", hex, err)
	}

	return id
}
