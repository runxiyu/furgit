package repository

import (
	"fmt"

	"codeberg.org/lindenii/furgit/config"
	objectid "codeberg.org/lindenii/furgit/object/id"
)

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
