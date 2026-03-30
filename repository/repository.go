// Package repository opens a typical on-disk Git repository and exposes its
// main stores and helpers.
//
// Start with [Open] when working with a bare repository root or a non-bare
// ".git" directory. [Repository] then provides access to ref storage, object
// storage, typed object fetching, commit queries, reachability helpers, and
// optional commit-graph access.
package repository

import (
	"os"

	"codeberg.org/lindenii/furgit/commitquery"
	"codeberg.org/lindenii/furgit/config"
	commitgraphread "codeberg.org/lindenii/furgit/format/commitgraph/read"
	"codeberg.org/lindenii/furgit/object/fetch"
	objectid "codeberg.org/lindenii/furgit/object/id"
	objectstore "codeberg.org/lindenii/furgit/object/store"
	objectloose "codeberg.org/lindenii/furgit/object/store/loose"
	objectpacked "codeberg.org/lindenii/furgit/object/store/packed"
	refstore "codeberg.org/lindenii/furgit/ref/store"
)

// Repository represents a typical on-disk Git repository by composing its
// stores and helpers together for access.
//
// Open expects a root for the Git directory itself:
// a bare repository root or a non-bare ".git" directory.
//
// Labels: MT-Safe, Close-Caller.
type Repository struct {
	config *config.Config
	algo   objectid.Algorithm

	objects         objectstore.Reader
	fetcher         *fetch.Fetcher
	objectsRoot     *os.Root
	objectsPackRoot *os.Root
	objectsLoose    *objectloose.Store
	objectsPacked   *objectpacked.Store
	commitGraph     *commitgraphread.Reader
	commitQueries   *commitquery.Queries
	refRoot         *os.Root
	refs            interface {
		refstore.ReadingStore
		refstore.TransactionalStore
		refstore.BatchStore
	}
}
