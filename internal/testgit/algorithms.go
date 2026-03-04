package testgit

import (
	"testing"

	"codeberg.org/lindenii/furgit/objectid"
)

// ForEachAlgorithm runs a subtest for every supported algorithm.
func ForEachAlgorithm(t *testing.T, fn func(t *testing.T, algo objectid.Algorithm)) {
	t.Helper()

	for _, algo := range objectid.SupportedAlgorithms() {
		t.Run(algo.String(), func(t *testing.T) {
			fn(t, algo)
		})
	}
}
