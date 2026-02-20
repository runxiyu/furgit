package testgit

import (
	"testing"

	"codeberg.org/lindenii/furgit/objectid"
)

// SupportedAlgorithms returns all object ID algorithms supported by furgit.
func SupportedAlgorithms() []objectid.Algorithm {
	return []objectid.Algorithm{
		objectid.AlgorithmSHA1,
		objectid.AlgorithmSHA256,
	}
}

// ForEachAlgorithm runs a subtest for every supported algorithm.
func ForEachAlgorithm(t *testing.T, fn func(t *testing.T, algo objectid.Algorithm)) {
	t.Helper()
	for _, algo := range SupportedAlgorithms() {
		t.Run(algo.String(), func(t *testing.T) {
			fn(t, algo)
		})
	}
}
