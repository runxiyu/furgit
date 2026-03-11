package commitquery

import (
	"slices"
)

// removeRedundant removes redundant merge-base candidates.
func removeRedundant(query *Query, candidates []nodeIndex) ([]nodeIndex, error) {
	for _, idx := range candidates {
		if query.effectiveGeneration(idx) != generationInfinity {
			return removeRedundantWithGen(query, candidates), nil
		}
	}

	return removeRedundantNoGen(query, candidates)
}

func removeRedundantNoGen(query *Query, candidates []nodeIndex) ([]nodeIndex, error) {
	redundant := make([]bool, len(candidates))
	work := make([]nodeIndex, 0, len(candidates)-1)
	filledIndex := make([]int, 0, len(candidates)-1)

	for i, candidate := range candidates {
		if redundant[i] {
			continue
		}

		work = work[:0]
		filledIndex = filledIndex[:0]

		minGeneration := query.effectiveGeneration(candidate)

		for j, other := range candidates {
			if i == j || redundant[j] {
				continue
			}

			work = append(work, other)
			filledIndex = append(filledIndex, j)

			otherGeneration := query.effectiveGeneration(other)
			if otherGeneration < minGeneration {
				minGeneration = otherGeneration
			}
		}

		err := query.paintDownToCommon(candidate, work, minGeneration)
		if err != nil {
			return nil, err
		}

		if query.hasAnyMarks(candidate, markRight) {
			redundant[i] = true
		}

		for j, other := range work {
			if query.hasAnyMarks(other, markLeft) {
				redundant[filledIndex[j]] = true
			}
		}

		query.clearTouchedMarks(allMarks)
	}

	out := make([]nodeIndex, 0, len(candidates))
	for i, idx := range candidates {
		if !redundant[i] {
			out = append(out, idx)
		}
	}

	return out, nil
}

func removeRedundantWithGen(query *Query, candidates []nodeIndex) []nodeIndex {
	sorted := append([]nodeIndex(nil), candidates...)
	slices.SortFunc(sorted, compareByGeneration(query))

	minGeneration := query.effectiveGeneration(sorted[0])
	minGenPos := 0
	countStillIndependent := len(candidates)

	query.beginMarkPhase()

	walkStart := make([]nodeIndex, 0, len(candidates)*2)

	for _, idx := range candidates {
		query.setMarks(idx, markResult)

		for _, parent := range query.parents(idx) {
			if query.hasAnyMarks(parent, markStale) {
				continue
			}

			query.setMarks(parent, markStale)
			walkStart = append(walkStart, parent)
		}
	}

	slices.SortFunc(walkStart, compareByGeneration(query))

	for _, idx := range walkStart {
		query.clearMarks(idx, markStale)
	}

	for i := len(walkStart) - 1; i >= 0 && countStillIndependent > 1; i-- {
		stack := []nodeIndex{walkStart[i]}
		query.setMarks(walkStart[i], markStale)

		for len(stack) > 0 {
			top := stack[len(stack)-1]

			if query.hasAnyMarks(top, markResult) {
				query.clearMarks(top, markResult)

				countStillIndependent--
				if countStillIndependent <= 1 {
					break
				}

				if top == sorted[minGenPos] {
					for minGenPos < len(sorted)-1 && query.hasAnyMarks(sorted[minGenPos], markStale) {
						minGenPos++
					}

					minGeneration = query.effectiveGeneration(sorted[minGenPos])
				}
			}

			if query.effectiveGeneration(top) < minGeneration {
				stack = stack[:len(stack)-1]

				continue
			}

			pushed := false

			for _, parent := range query.parents(top) {
				if query.hasAnyMarks(parent, markStale) {
					continue
				}

				query.setMarks(parent, markStale)
				stack = append(stack, parent)
				pushed = true

				break
			}

			if !pushed {
				stack = stack[:len(stack)-1]
			}
		}
	}

	out := make([]nodeIndex, 0, len(candidates))
	for _, idx := range candidates {
		if !query.hasAnyMarks(idx, markStale) {
			out = append(out, idx)
		}
	}

	query.clearTouchedMarks(markStale | markResult)

	return out
}
