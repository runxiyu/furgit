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
	seq := query.Seq()

	var (
		first objectid.ObjectID
		ok    bool
	)

	seq(func(id objectid.ObjectID) bool {
		first = id
		ok = true

		return false
	})

	err := query.Err()
	if err != nil {
		return objectid.ObjectID{}, false, err
	}

	if !ok {
		return objectid.ObjectID{}, false, nil
	}

	return first, true, nil
}
