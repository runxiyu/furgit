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

	"lindenii.org/go/furgit/commitquery"
	"lindenii.org/go/furgit/config"
	commitgraphread "lindenii.org/go/furgit/format/commitgraph/read"
	"lindenii.org/go/furgit/object/fetch"
	objectid "lindenii.org/go/furgit/object/id"
	objectdual "lindenii.org/go/furgit/object/store/dual"
	objectloose "lindenii.org/go/furgit/object/store/loose"
	objectpacked "lindenii.org/go/furgit/object/store/packed"
	refstore "lindenii.org/go/furgit/ref/store"
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

	objects         *objectdual.Dual
	fetcher         *fetch.Fetcher
	objectsRoot     *os.Root
	objectsPackRoot *os.Root
	objectsLoose    *objectloose.Store
	objectsPacked   *objectpacked.Store
	commitGraph     *commitgraphread.Reader
	commitQueries   *commitquery.Queries
	refRoot         *os.Root
	refs            interface {
		refstore.Reader
		refstore.Transactioner
		refstore.Batcher
	}
}
