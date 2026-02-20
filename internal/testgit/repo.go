package testgit

import "codeberg.org/lindenii/furgit/objectid"

// TestRepo is a temporary git repository harness for integration tests.
type TestRepo struct {
	dir  string
	algo objectid.Algorithm
	env  []string
}
