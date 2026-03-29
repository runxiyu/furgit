package commitquery

import "slices"

// mergeBases returns internal merge-base candidates for two resolved nodes.
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
