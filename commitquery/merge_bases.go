package commitquery

import (
	"slices"

	objectid "codeberg.org/lindenii/furgit/object/id"
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

// MergeBase reports one merge base between left and right, if any.
func (query *query) MergeBase(left, right objectid.ObjectID) (objectid.ObjectID, bool, error) {
	bases, err := query.MergeBases(left, right)
	if err != nil {
		return objectid.ObjectID{}, false, err
	}

	if len(bases) == 0 {
		return objectid.ObjectID{}, false, nil
	}

	return bases[0], true, nil
}

func (query *query) mergeBases(left, right nodeIndex) ([]nodeIndex, error) {
	if left == right {
		return []nodeIndex{left}, nil
	}

	err := query.paintDownToCommon(left, []nodeIndex{right}, 0)
	if err != nil {
		return nil, err
	}

	candidates := query.collectMarkedResults()

	if len(candidates) <= 1 {
		slices.SortFunc(candidates, query.compare)

		return candidates, nil
	}

	query.clearTouchedMarks(allMarks)

	reduced, err := removeRedundant(query, candidates)
	if err != nil {
		return nil, err
	}

	slices.SortFunc(reduced, query.compare)

	return reduced, nil
}
