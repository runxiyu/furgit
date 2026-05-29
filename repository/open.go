package repository

import (
	"fmt"
	"os"

	"lindenii.org/go/furgit/commitquery"
	"lindenii.org/go/furgit/object/fetch"
	reffiles "lindenii.org/go/furgit/ref/store/files"
)

// Open opens a repository and wires its stores and helpers from the on-disk
// repository format.
//
// root must refer to the Git directory itself:
// a bare repository root or a non-bare ".git" directory.
//
// Labels: Deps-Borrowed.
func Open(root *os.Root) (repo *Repository, err error) {
	repo = &Repository{}

	defer func() {
		if err != nil {
			_ = repo.Close()
		}
	}()

	cfg, err := parseRepositoryConfig(root)
	if err != nil {
		return repo, err
	}

	repo.config = cfg

	algo, err := detectObjectAlgorithm(cfg)
	if err != nil {
		return repo, err
	}

	repo.algo = algo

	objects, objectsRoot, objectsPackRoot, objectsLoose, objectsPacked, err := openObjectStore(root, algo)
	if err != nil {
		return repo, err
	}

	repo.objects = objects
	repo.fetcher = fetch.New(objects)
	repo.objectsRoot = objectsRoot
	repo.objectsPackRoot = objectsPackRoot
	repo.objectsLoose = objectsLoose
	repo.objectsPacked = objectsPacked

	commitGraph, err := openCommitGraph(objectsRoot, algo)
	if err != nil {
		return repo, err
	}

	repo.commitGraph = commitGraph
	repo.commitQueries = commitquery.New(repo.fetcher, commitGraph)

	refRoot, err := root.OpenRoot(".")
	if err != nil {
		return repo, fmt.Errorf("repository: open root for refs: %w", err)
	}

	refs, err := reffiles.New(refRoot, algo, detectPackedRefsTimeout(cfg))
	if err != nil {
		_ = refRoot.Close()

		return repo, err
	}

	repo.refs = refs
	repo.refRoot = refRoot

	return repo, nil
}
