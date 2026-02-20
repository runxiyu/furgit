package testgit

import (
	"testing"

	"codeberg.org/lindenii/furgit/objectid"
)

// RevParse resolves rev expressions to object IDs.
func (repo *TestRepo) RevParse(tb testing.TB, spec string) objectid.ObjectID {
	tb.Helper()
	hex := repo.Run(tb, "rev-parse", spec)
	id, err := objectid.ParseHex(repo.algo, hex)
	if err != nil {
		tb.Fatalf("parse rev-parse output %q: %v", hex, err)
	}
	return id
}
