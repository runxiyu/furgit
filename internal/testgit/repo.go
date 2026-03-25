// Package testgit provides helpers for integration tests with upstream git(1).
package testgit

import objectid "codeberg.org/lindenii/furgit/object/id"

// TestRepo is a temporary git repository harness for integration tests.
type TestRepo struct {
	dir  string
	algo objectid.Algorithm
	env  []string
}
