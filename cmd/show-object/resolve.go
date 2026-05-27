package main

import (
	"strings"

	objectid "lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/repository"
)

func resolveInput(repo *repository.Repository, input string) (objectid.ObjectID, error) {
	id, err := objectid.ParseHex(repo.Algorithm(), strings.TrimSpace(input))
	if err == nil {
		return id, nil
	}

	resolved, err := repo.RefStore().ResolveToDetached(input)
	if err != nil {
		return objectid.ObjectID{}, err
	}

	return resolved.ID, nil
}
