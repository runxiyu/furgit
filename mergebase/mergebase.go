// Package mergebase computes best common ancestors between commits.
package mergebase

import (
	commitgraphread "codeberg.org/lindenii/furgit/format/commitgraph/read"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore"
)

// Bases is one iterator merge-base query.
type Bases struct {
	store objectstore.Store
	graph *commitgraphread.Reader
	left  objectid.ObjectID
	right objectid.ObjectID

	seqUsed bool
	err     error
}
