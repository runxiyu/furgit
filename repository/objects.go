package repository

import (
	"fmt"
	"os"

	objectid "lindenii.org/go/furgit/object/id"
	objectstore "lindenii.org/go/furgit/object/store"
	objectdual "lindenii.org/go/furgit/object/store/dual"
	objectloose "lindenii.org/go/furgit/object/store/loose"
	objectpacked "lindenii.org/go/furgit/object/store/packed"
)

// openObjectStore opens the roots and object stores of both
// the loose and packed stores for a particular repo root and
// object ID hashing algorithm.
//
// Since real object store implementations do not take ownership of
// the roots given to them, and since composite object stores do not
// take ownership of the object stores that they consist of, all
// of them are returned and should be closed by the caller.
//
//nolint:ireturn
func openObjectStore(
	root *os.Root,
	algo objectid.Algorithm,
) (
	objects *objectdual.Dual,
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

	err = objectsRoot.Mkdir("pack", 0o755)
	if err != nil && !os.IsExist(err) {
		_ = objectsLoose.Close()
		_ = objectsRoot.Close()

		return nil, nil, nil, nil, nil, fmt.Errorf("repository: create objects/pack: %w", err)
	}

	objectsPackRoot, err = objectsRoot.OpenRoot("pack")
	if err != nil {
		_ = objectsLoose.Close()
		_ = objectsRoot.Close()

		return nil, nil, nil, nil, nil, fmt.Errorf("repository: open objects/pack: %w", err)
	}

	objectsPacked, err = objectpacked.New(
		objectsPackRoot,
		algo,
		objectpacked.Options{
			RefreshPolicy: objectpacked.RefreshPolicyNever,
			WriteRev:      true,
		},
	)
	if err != nil {
		_ = objectsPackRoot.Close()
		_ = objectsLoose.Close()
		_ = objectsRoot.Close()

		return nil, nil, nil, nil, nil, err
	}

	objects = objectdual.New(objectsLoose, objectsPacked)

	return objects, objectsRoot, objectsPackRoot, objectsLoose, objectsPacked, nil
}

// ObjectStore returns the configured object store.
//
// Use ObjectStore for direct object-ID lookups, object headers, sizes, raw object
// bytes, streamed object contents, object writes, pack ingestion, and
// coordinated quarantines. Callers who want typed object values should usually
// prefer [Repository.Fetcher].
//
// Labels: Life-Parent.
//
//nolint:ireturn
func (repo *Repository) ObjectStore() interface {
	objectstore.Reader
	objectstore.Writer
	objectstore.Quarantiner
} {
	return repo.objects
}
