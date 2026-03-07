package repository

import (
	"fmt"
	"os"
	"time"

	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/refstore"
	reffiles "codeberg.org/lindenii/furgit/refstore/files"
)

//nolint:ireturn
func openRefStore(root *os.Root, algo objectid.Algorithm, packedRefsTimeout time.Duration) (out refstore.ReadWriteStore, err error) {
	refRoot, err := root.OpenRoot(".")
	if err != nil {
		return nil, fmt.Errorf("repository: open root for refs: %w", err)
	}

	store, err := reffiles.New(refRoot, algo, packedRefsTimeout)
	if err != nil {
		_ = refRoot.Close()

		return nil, err
	}

	return store, nil
}

// Refs returns the configured ref store.
//
//nolint:ireturn
func (repo *Repository) Refs() refstore.ReadWriteStore {
	return repo.refs
}
