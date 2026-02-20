package testgit

import "codeberg.org/lindenii/furgit/oid"

// TestRepo is a temporary git repository harness for integration tests.
type TestRepo struct {
	dir  string
	algo oid.Algorithm
	env  []string
}
