package main

import (
	"strings"

	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/repository"
)

func resolveInput(repo *repository.Repository, input string) (objectid.ObjectID, error) {
	id, err := objectid.ParseHex(repo.Algorithm(), strings.TrimSpace(input))
	if err == nil {
		return id, nil
	}

	resolved, err := repo.Refs().ResolveToDetached(input)
	if err != nil {
		return objectid.ObjectID{}, err
	}

	return resolved.ID, nil
}
