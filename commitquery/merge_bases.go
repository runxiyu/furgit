package commitquery

import (
	"slices"

	"codeberg.org/lindenii/furgit/objectid"
)

// MergeBases reports all merge bases in Git's merge-base --all order.
//
// Both inputs are peeled through annotated tags before commit traversal.
func (query *Query) MergeBases(left, right objectid.ObjectID) ([]objectid.ObjectID, error) {
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
func (query *Query) MergeBase(left, right objectid.ObjectID) (objectid.ObjectID, bool, error) {
	bases, err := query.MergeBases(left, right)
	if err != nil {
		return objectid.ObjectID{}, false, err
	}

	if len(bases) == 0 {
		return objectid.ObjectID{}, false, nil
	}

	return bases[0], true, nil
}

func (query *Query) mergeBases(left, right nodeIndex) ([]nodeIndex, error) {
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

func (query *Query) paintDownToCommon(left nodeIndex, rights []nodeIndex, minGeneration uint64) error {
	query.beginMarkPhase()

	query.setMarks(left, markLeft)

	if len(rights) == 0 {
		query.setMarks(left, markResult)

		return nil
	}

	queue := newPriorityQueue(query)
	queue.PushNode(left)

	for _, right := range rights {
		query.setMarks(right, markRight)
		queue.PushNode(right)
	}

	lastGeneration := generationInfinity

	for query.queueHasNonStale(queue) {
		idx := queue.PopNode()

		generation := query.effectiveGeneration(idx)
		if generation > lastGeneration {
			return errBadGenerationOrder
		}

		lastGeneration = generation
		if generation < minGeneration {
			break
		}

		flags := query.marks(idx) & (markLeft | markRight | markStale)
		if flags == (markLeft | markRight) {
			query.setMarks(idx, markResult)

			flags |= markStale
		}

		for _, parent := range query.parents(idx) {
			if query.hasAllMarks(parent, flags) {
				continue
			}

			query.setMarks(parent, flags)
			queue.PushNode(parent)
		}
	}

	return nil
}

func (query *Query) queueHasNonStale(queue *priorityQueue) bool {
	for _, idx := range queue.items {
		if !query.hasAnyMarks(idx, markStale) {
			return true
		}
	}

	return false
}

func (query *Query) collectMarkedResults() []nodeIndex {
	out := make([]nodeIndex, 0, 4)

	for _, idx := range query.touched {
		if !query.hasAnyMarks(idx, markResult) {
			continue
		}

		if query.hasAnyMarks(idx, markStale) {
			continue
		}

		out = append(out, idx)
	}

	return out
}
