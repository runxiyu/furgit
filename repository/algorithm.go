package repository

import (
	"fmt"

	"lindenii.org/go/furgit/config"
	objectid "lindenii.org/go/furgit/object/id"
)

// detectObjectAlgorithm uses a repository's configuration to detect
// the expected Object ID hashing algorithm.
func detectObjectAlgorithm(cfg *config.Config) (objectid.Algorithm, error) {
	algoName := cfg.Lookup("extensions", "", "objectformat").Value
	if algoName == "" {
		algoName = objectid.AlgorithmSHA1.String()
	}

	algo, ok := objectid.ParseAlgorithm(algoName)
	if !ok {
		return objectid.AlgorithmUnknown, fmt.Errorf("repository: unsupported object format %q", algoName)
	}

	return algo, nil
}

// Algorithm returns the repository object ID algorithm.
func (repo *Repository) Algorithm() objectid.Algorithm {
	return repo.algo
}
