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
func openObjectStore(
	root *os.Root,
	algo objectid.Algorithm,
) (
	objects objectstore.Store,
	objectsRoot *os.Root,
	objectsPackRoot *os.Root,
	objectsLooseForWritingOnly *objectloose.Store,
	objectsWriteRoot *os.Root,
	err error,
) {
	objectsRoot, err = root.OpenRoot("objects")
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("repository: open objects: %w", err)
	}

	looseStore, err := objectloose.New(objectsRoot, algo)
	if err != nil {
		_ = objectsRoot.Close()

		return nil, nil, nil, nil, nil, err
	}

	backends := []objectstore.Store{looseStore}
	objectsPackRoot, err = objectsRoot.OpenRoot("pack")

	if err == nil {
		var packedStore *objectpacked.Store

		packedStore, err = objectpacked.New(
			objectsPackRoot,
			algo,
			objectpacked.Options{RefreshPolicy: objectpacked.RefreshPolicyNever},
		)
		if err != nil {
			_ = objectsPackRoot.Close()
			_ = looseStore.Close()
			_ = objectsRoot.Close()

			return nil, nil, nil, nil, nil, err
		}

		backends = append(backends, packedStore)
	} else if !errors.Is(err, os.ErrNotExist) {
		_ = looseStore.Close()
		_ = objectsRoot.Close()

		return nil, nil, nil, nil, nil, fmt.Errorf("repository: open objects/pack: %w", err)
	}

	objects = objectmix.New(backends...)

	objectsWriteRoot, err = root.OpenRoot("objects")
	if err != nil {
		_ = objects.Close()
		if objectsPackRoot != nil {
			_ = objectsPackRoot.Close()
		}
		_ = objectsRoot.Close()

		return nil, nil, nil, nil, nil, fmt.Errorf("repository: open objects for loose writing: %w", err)
	}

	objectsLooseForWritingOnly, err = objectloose.New(objectsWriteRoot, algo)
	if err != nil {
		_ = objects.Close()
		_ = objectsWriteRoot.Close()
		if objectsPackRoot != nil {
			_ = objectsPackRoot.Close()
		}
		_ = objectsRoot.Close()

		return nil, nil, nil, nil, nil, err
	}

	return objects, objectsRoot, objectsPackRoot, objectsLooseForWritingOnly, objectsWriteRoot, nil
}

// Objects returns the configured object store.
//
//nolint:ireturn
func (repo *Repository) Objects() objectstore.Store {
	return repo.objects
}
