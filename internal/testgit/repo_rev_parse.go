package testgit

import (
	"testing"

	"codeberg.org/lindenii/furgit/objectid"
)

// RevParse resolves rev expressions to object IDs.
func (testRepo *TestRepo) RevParse(tb testing.TB, spec string) objectid.ObjectID {
	tb.Helper()
	hex := testRepo.Run(tb, "rev-parse", spec)
	id, err := objectid.ParseHex(testRepo.algo, hex)
	if err != nil {
		tb.Fatalf("parse rev-parse output %q: %v", hex, err)
	}
	return id
}
