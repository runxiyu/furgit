// Package repository wires object and ref storage for one Git repository.
package repository

import (
	"errors"

	"codeberg.org/lindenii/furgit/config"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore"
	objectloose "codeberg.org/lindenii/furgit/objectstore/loose"
	"codeberg.org/lindenii/furgit/refstore"
)

// Repository is a thin composition root for repository-local stores.
//
// Open expects a root for the Git directory itself:
// a bare repository root or a non-bare ".git" directory.
type Repository struct {
	config *config.Config
	algo   objectid.Algorithm

	objects                    objectstore.Store
	objectsLooseForWritingOnly *objectloose.Store
	refs                       refstore.Store
}

// Algorithm returns the repository object ID algorithm.
func (repo *Repository) Algorithm() objectid.Algorithm {
	return repo.algo
}

// Config returns the parsed repository configuration snapshot.
//
// The returned pointer is owned by Repository. Callers should treat it as
// read-only.
func (repo *Repository) Config() *config.Config {
	return repo.config
}

// Objects returns the configured object store.
func (repo *Repository) Objects() objectstore.Store {
	return repo.objects
}

// Refs returns the configured ref store.
func (repo *Repository) Refs() refstore.Store {
	return repo.refs
}

// Close closes owned stores and filesystem roots.
// The behavior of the repo after Close is undefined.
func (repo *Repository) Close() error {
	var errs []error

	if repo.refs != nil {
		err := repo.refs.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}

	if repo.objects != nil {
		err := repo.objects.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}

	if repo.objectsLooseForWritingOnly != nil {
		err := repo.objectsLooseForWritingOnly.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
