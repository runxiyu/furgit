package repository

import (
	"errors"
	"fmt"
	"os"

	objectid "codeberg.org/lindenii/furgit/object/id"
	objectstorer "codeberg.org/lindenii/furgit/object/storer"
	objectloose "codeberg.org/lindenii/furgit/object/storer/loose"
	objectmix "codeberg.org/lindenii/furgit/object/storer/mix"
	objectpacked "codeberg.org/lindenii/furgit/object/storer/packed"
)

//nolint:ireturn
func openObjectStore(
	root *os.Root,
	algo objectid.Algorithm,
) (
	objects objectstorer.Store,
	objectsRoot *os.Root,
	objectsPackRoot *os.Root,
	objectsLoose *objectloose.Store,
	objectsPacked *objectpacked.Store,
	err error,
) {
	objectsRoot, err = root.OpenRoot("objects")
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("repository: open objects: %w", err)
	}

	objectsLoose, err = objectloose.New(objectsRoot, algo)
	if err != nil {
		_ = objectsRoot.Close()

		return nil, nil, nil, nil, nil, err
	}

	backends := []objectstorer.Store{objectsLoose}

	objectsPackRoot, err = objectsRoot.OpenRoot("pack")
	if err == nil {
		objectsPacked, err = objectpacked.New(
			objectsPackRoot,
			algo,
			objectpacked.Options{RefreshPolicy: objectpacked.RefreshPolicyNever},
		)
		if err != nil {
			_ = objectsPackRoot.Close()
			_ = objectsLoose.Close()
			_ = objectsRoot.Close()

			return nil, nil, nil, nil, nil, err
		}

		backends = append(backends, objectsPacked)
	} else if !errors.Is(err, os.ErrNotExist) {
		_ = objectsLoose.Close()
		_ = objectsRoot.Close()

		return nil, nil, nil, nil, nil, fmt.Errorf("repository: open objects/pack: %w", err)
	}

	objects = objectmix.New(backends...)

	return objects, objectsRoot, objectsPackRoot, objectsLoose, objectsPacked, nil
}

// Objects returns the configured object store.
//
// The returned store is owned by Repository and borrows repository-managed
// resources. Callers must not close it directly, and it must not be used after
// Close.
//
//nolint:ireturn
func (repo *Repository) Objects() objectstorer.Store {
	return repo.objects
}
