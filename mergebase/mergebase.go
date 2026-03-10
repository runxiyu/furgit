// Package mergebase computes best common ancestors between commits.
package mergebase

import (
	commitgraphread "codeberg.org/lindenii/furgit/commitgraph/read"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore"
)

// Bases is one merge-base query over two commit roots.
type Bases struct {
	store objectstore.Store
	graph *commitgraphread.Reader
	left  objectid.ObjectID
	right objectid.ObjectID

	computed bool
	bases    []objectid.ObjectID
	err      error
}
