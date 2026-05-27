package commitquery

import (
	"slices"

	objectid "lindenii.org/go/furgit/object/id"
)

// MergeBases reports all merge bases in Git's merge-base --all order.
//
// Both inputs are peeled through annotated tags before commit traversal.
func (query *query) MergeBases(left, right objectid.ObjectID) ([]objectid.ObjectID, error) {
	leftIdx, err := query.resolveCommitish(left)
	if err != nil {
		return nil, err
	}

	rightIdx, err := query.resolveCommitish(right)
	if err != nil {
		return nil, err
	}

	candidates, err := query.mergeBases(leftIdx, rightIdx)
	if err != nil {
		return nil, err
	}

	slices.SortFunc(candidates, func(left, right nodeIndex) int {
		switch {
		case query.commitTime(left) > query.commitTime(right):
			return -1
		case query.commitTime(left) < query.commitTime(right):
			return 1
		default:
			return objectid.Compare(query.id(left), query.id(right))
		}
	})

	out := make([]objectid.ObjectID, 0, len(candidates))
	for _, idx := range candidates {
		out = append(out, query.id(idx))
	}

	return out, nil
}
