// Package repository wires object and ref storage for one Git repository.
package repository

import (
	"os"

	"codeberg.org/lindenii/furgit/config"
	objectid "codeberg.org/lindenii/furgit/object/id"
	"codeberg.org/lindenii/furgit/objectstore"
	objectloose "codeberg.org/lindenii/furgit/objectstore/loose"
	objectpacked "codeberg.org/lindenii/furgit/objectstore/packed"
	"codeberg.org/lindenii/furgit/refstore"
)

// Repository is a thin composition root for repository-local stores.
//
// Open expects a root for the Git directory itself:
// a bare repository root or a non-bare ".git" directory.
//
// Accessors such as Objects, Refs, Resolver, and LooseStoreForWriting return
// views backed by resources owned by Repository. Those values borrow the
// repository's stores and filesystem roots and must not be used after Close.
type Repository struct {
	config *config.Config
	algo   objectid.Algorithm

	objects         objectstore.Store
	objectsRoot     *os.Root
	objectsPackRoot *os.Root
	objectsLoose    *objectloose.Store
	objectsPacked   *objectpacked.Store
	refRoot         *os.Root
	refs            refstore.ReadWriteStore
}
