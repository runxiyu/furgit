package commitquery

import "slices"

// MergeBases computes fully reduced merge bases using one query context.
func MergeBases(ctx *Context, left, right NodeIndex) ([]NodeIndex, error) {
	if left == right {
		return []NodeIndex{left}, nil
	}

	err := paintDownToCommon(ctx, left, []NodeIndex{right}, 0)
	if err != nil {
		return nil, err
	}

	candidates := collectMarkedResults(ctx)

	if len(candidates) <= 1 {
		slices.SortFunc(candidates, ctx.Compare)

		return candidates, nil
	}

	ctx.ClearTouchedMarks(allMarks)

	reduced, err := removeRedundant(ctx, candidates)
	if err != nil {
		return nil, err
	}

	slices.SortFunc(reduced, ctx.Compare)

	return reduced, nil
}

func paintDownToCommon(ctx *Context, left NodeIndex, rights []NodeIndex, minGeneration uint64) error {
	ctx.BeginMarkPhase()

	ctx.SetMarks(left, markLeft)

	if len(rights) == 0 {
		ctx.SetMarks(left, markResult)

		return nil
	}

	queue := newPriorityQueue(ctx)
	queue.PushNode(left)

	for _, right := range rights {
		ctx.SetMarks(right, markRight)
		queue.PushNode(right)
	}

	lastGeneration := generationInfinity

	for queueHasNonStale(ctx, queue) {
		idx := queue.PopNode()

		generation := ctx.EffectiveGeneration(idx)
		if generation > lastGeneration {
			return errBadGenerationOrder
		}

		lastGeneration = generation
		if generation < minGeneration {
			break
		}

		flags := ctx.Marks(idx) & (markLeft | markRight | markStale)
		if flags == (markLeft | markRight) {
			ctx.SetMarks(idx, markResult)

			flags |= markStale
		}

		for _, parent := range ctx.Parents(idx) {
			if ctx.HasAllMarks(parent, flags) {
				continue
			}

			ctx.SetMarks(parent, flags)
			queue.PushNode(parent)
		}
	}

	return nil
}

func queueHasNonStale(ctx *Context, queue *priorityQueue) bool {
	for _, idx := range queue.items {
		if !ctx.HasAnyMarks(idx, markStale) {
			return true
		}
	}

	return false
}

func collectMarkedResults(ctx *Context) []NodeIndex {
	out := make([]NodeIndex, 0, 4)

	for _, idx := range ctx.touched {
		if !ctx.HasAnyMarks(idx, markResult) {
			continue
		}

		if ctx.HasAnyMarks(idx, markStale) {
			continue
		}

		out = append(out, idx)
	}

	return out
}
