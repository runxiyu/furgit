package repository

import "os"

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

	objects, objectsRoot, objectsPackRoot, objectsLooseForWritingOnly, objectsWriteRoot, err := openObjectStore(root, algo)
	if err != nil {
		return nil, err
	}

	repo.objects = objects
	repo.objectsRoot = objectsRoot
	repo.objectsPackRoot = objectsPackRoot
	repo.objectsLooseForWritingOnly = objectsLooseForWritingOnly
	repo.objectsWriteRoot = objectsWriteRoot

	refs, err := openRefStore(root, algo, detectPackedRefsTimeout(cfg))
	if err != nil {
		return nil, err
	}

	repo.refs = refs

	return repo, nil
}
