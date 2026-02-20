package testgit

import (
	"testing"

	"codeberg.org/lindenii/furgit/oid"
)

// SupportedAlgorithms returns all object ID algorithms supported by furgit.
func SupportedAlgorithms() []oid.Algorithm {
	return []oid.Algorithm{
		oid.AlgorithmSHA1,
		oid.AlgorithmSHA256,
	}
}

// ForEachAlgorithm runs a subtest for every supported algorithm.
func ForEachAlgorithm(t *testing.T, fn func(t *testing.T, algo oid.Algorithm)) {
	t.Helper()
	for _, algo := range SupportedAlgorithms() {
		t.Run(algo.String(), func(t *testing.T) {
			fn(t, algo)
		})
	}
}
