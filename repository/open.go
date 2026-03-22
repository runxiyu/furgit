package repository

import (
	"fmt"
	"os"

	reffiles "codeberg.org/lindenii/furgit/refstore/files"
)

// Open opens a repository and wires object/ref stores from its on-disk format.
//
// Open borrows root during construction and does not close it.
func Open(root *os.Root) (repo *Repository, err error) {
	repo = &Repository{}

	defer func() {
		if err != nil {
			_ = repo.Close()
		}
	}()

	cfg, err := parseRepositoryConfig(root)
	if err != nil {
		return nil, err
	}

	repo.config = cfg

	algo, err := detectObjectAlgorithm(cfg)
	if err != nil {
		return nil, err
	}

	repo.algo = algo

	objects, objectsRoot, objectsPackRoot, objectsLoose, objectsPacked, err := openObjectStore(root, algo)
	if err != nil {
		return nil, err
	}

	repo.objects = objects
	repo.objectsRoot = objectsRoot
	repo.objectsPackRoot = objectsPackRoot
	repo.objectsLoose = objectsLoose
	repo.objectsPacked = objectsPacked

	refRoot, err := root.OpenRoot(".")
	if err != nil {
		return nil, fmt.Errorf("repository: open root for refs: %w", err)
	}

	refs, err := reffiles.New(refRoot, algo, detectPackedRefsTimeout(cfg))
	if err != nil {
		_ = refRoot.Close()

		return nil, err
	}

	repo.refs = refs
	repo.refRoot = refRoot

	return repo, nil
}
