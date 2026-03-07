package mergebase

import (
	commitgraphread "codeberg.org/lindenii/furgit/format/commitgraph/read"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objectstore"
)

// Base reports one merge base between left and right, if any.
//
// Both inputs are peeled through annotated tags before commit traversal.
func Base(
	store objectstore.Store,
	graph *commitgraphread.Reader,
	left objectid.ObjectID,
	right objectid.ObjectID,
) (objectid.ObjectID, bool, error) {
	query := Query(store, graph, left, right)

	bases, err := query.All()
	if err != nil {
		return objectid.ObjectID{}, false, err
	}

	if len(bases) == 0 {
		return objectid.ObjectID{}, false, nil
	}

	return bases[0], true, nil
}
