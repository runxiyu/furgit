package commitquery

func (query *query) paintDownToCommon(left nodeIndex, rights []nodeIndex, minGeneration uint64) error {
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
