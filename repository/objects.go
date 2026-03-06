package repository

import (
	"errors"
	"fmt"
	"os"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore"
	objectloose "codeberg.org/lindenii/furgit/objectstore/loose"
	objectmix "codeberg.org/lindenii/furgit/objectstore/mix"
	objectpacked "codeberg.org/lindenii/furgit/objectstore/packed"
)

//nolint:ireturn
func openObjectStore(root *os.Root, algo objectid.Algorithm) (objectstore.Store, *objectloose.Store, error) {
	objectsRoot, err := root.OpenRoot("objects")
	if err != nil {
		return nil, nil, fmt.Errorf("repository: open objects: %w", err)
	}

	looseStore, err := objectloose.New(objectsRoot, algo)
	if err != nil {
		return nil, nil, err
	}

	backends := []objectstore.Store{looseStore}

	packRoot, err := objectsRoot.OpenRoot("pack")
	if err == nil {
		var packedStore *objectpacked.Store

		packedStore, err = objectpacked.New(packRoot, algo)
		if err != nil {
			_ = looseStore.Close()

			return nil, nil, err
		}

		backends = append(backends, packedStore)
	} else if !errors.Is(err, os.ErrNotExist) {
		_ = looseStore.Close()

		return nil, nil, fmt.Errorf("repository: open objects/pack: %w", err)
	}

	objectsChain := objectmix.New(backends...)

	objectsRootForWriting, err := root.OpenRoot("objects")
	if err != nil {
		_ = objectsChain.Close()

		return nil, nil, fmt.Errorf("repository: open objects for loose writing: %w", err)
	}

	objectsLooseForWritingOnly, err := objectloose.New(objectsRootForWriting, algo)
	if err != nil {
		_ = objectsRootForWriting.Close()
		_ = objectsChain.Close()

		return nil, nil, err
	}

	return objectsChain, objectsLooseForWritingOnly, nil
}

// Objects returns the configured object store.
//
//nolint:ireturn
func (repo *Repository) Objects() objectstore.Store {
	return repo.objects
}
