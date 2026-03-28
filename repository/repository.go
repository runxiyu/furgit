// Package repository opens stores and other objects to access a typical on-disk repo.
package repository

import (
	"os"

	"codeberg.org/lindenii/furgit/config"
	objectid "codeberg.org/lindenii/furgit/object/id"
	objectstore "codeberg.org/lindenii/furgit/object/store"
	objectloose "codeberg.org/lindenii/furgit/object/store/loose"
	objectpacked "codeberg.org/lindenii/furgit/object/store/packed"
	refstore "codeberg.org/lindenii/furgit/ref/store"
)

// Repository represents a typical on-disk Git repository by composing
// its stores together for access.
//
// Open expects a root for the Git directory itself:
// a bare repository root or a non-bare ".git" directory.
//
// Accessors such as [Repository.Objects], [Repository.Refs],
// [Repository.Fetcher], and [Repository.LooseStoreForWriting] return
// repository-backed views.
//
// Labels: MT-Safe, Close-Caller.
type Repository struct {
	config *config.Config
	algo   objectid.Algorithm

	objects         objectstore.ReadingStore
	objectsRoot     *os.Root
	objectsPackRoot *os.Root
	objectsLoose    *objectloose.Store
	objectsPacked   *objectpacked.Store
	refRoot         *os.Root
	refs            refstore.ReadWriteStore
}
