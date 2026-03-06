// Package ancestor answers commit ancestry queries.
package ancestor

import (
	commitgraphread "codeberg.org/lindenii/furgit/format/commitgraph/read"
	"codeberg.org/lindenii/furgit/internal/commitquery"
	"codeberg.org/lindenii/furgit/internal/peel"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore"
)

// Is reports whether ancestor is reachable from descendant through commit
// parent edges.
//
// Both inputs are peeled through annotated tags before commit traversal.
func Is(
	store objectstore.Store,
	graph *commitgraphread.Reader,
	ancestor objectid.ObjectID,
	descendant objectid.ObjectID,
) (bool, error) {
	ancestorCommit, err := peel.ToCommit(store, ancestor)
	if err != nil {
		return false, err
	}

	descendantCommit, err := peel.ToCommit(store, descendant)
	if err != nil {
		return false, err
	}

	ctx := commitquery.NewContext(store, graph)

	ancestorIdx, err := ctx.ResolveOID(ancestorCommit)
	if err != nil {
		return false, err
	}

	descendantIdx, err := ctx.ResolveOID(descendantCommit)
	if err != nil {
		return false, err
	}

	return commitquery.IsAncestor(ctx, ancestorIdx, descendantIdx)
}
