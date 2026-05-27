package commitquery

import "lindenii.org/go/furgit/internal/priorityqueue"

// paintDownToCommon propagates left and right marks downward until common nodes.
func (query *query) paintDownToCommon(left nodeIndex, rights []nodeIndex, minGeneration uint64) error {
	query.beginMarkPhase()

	query.setMarks(left, markLeft)

	if len(rights) == 0 {
		query.setMarks(left, markResult)

		return nil
	}

	queue := priorityqueue.New(func(left, right nodeIndex) bool {
		return query.compare(left, right) > 0
	})
	queue.Push(left)

	for _, right := range rights {
		query.setMarks(right, markRight)
		queue.Push(right)
	}

	lastGeneration := generationInfinity

	for queue.Len() > 0 {
		idx, ok := queue.Pop()
		if !ok {
			break
		}

		if query.hasAnyMarks(idx, markStale) {
			continue
		}

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
			queue.Push(parent)
		}
	}

	return nil
}
