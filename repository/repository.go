// Package repository wires object and ref storage for one Git repository.
package repository

import (
	"os"

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
	objectsRoot                *os.Root
	objectsPackRoot            *os.Root
	objectsLooseForWritingOnly *objectloose.Store
	objectsWriteRoot           *os.Root
	refRoot                    *os.Root
	refs                       refstore.ReadWriteStore
}
