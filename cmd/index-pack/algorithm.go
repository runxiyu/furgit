package main

import (
	"fmt"

	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/repository"
)

func resolveAlgorithm(repo *repository.Repository, objectFormat string) (objectid.Algorithm, error) {
	if objectFormat != "" {
		algo, ok := objectid.ParseAlgorithm(objectFormat)
		if !ok {
			return objectid.AlgorithmUnknown, fmt.Errorf("invalid object format %q", objectFormat)
		}

		return algo, nil
	}

	if repo != nil {
		return repo.Algorithm(), nil
	}

	return objectid.AlgorithmSHA1, nil
}
