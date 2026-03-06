package commitquery

import "slices"

// MergeBases computes fully reduced merge bases using one query context.
func MergeBases(ctx *Context, left, right NodeIndex) ([]NodeIndex, error) {
	if left == right {
		return []NodeIndex{left}, nil
	}

	candidates, err := paintDownToCommon(ctx, left, []NodeIndex{right}, 0)
	if err != nil {
		return nil, err
	}

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

func paintDownToCommon(ctx *Context, left NodeIndex, rights []NodeIndex, minGeneration uint64) ([]NodeIndex, error) {
	ctx.BeginMarkPhase()

	ctx.SetMarks(left, markLeft)

	if len(rights) == 0 {
		return []NodeIndex{left}, nil
	}

	queue := NewPriorityQueue(ctx)
	queue.PushNode(left)

	for _, right := range rights {
		ctx.SetMarks(right, markRight)
		queue.PushNode(right)
	}

	lastGeneration := generationInfinity
	results := make([]NodeIndex, 0, 4)

	for queueHasNonStale(ctx, queue) {
		idx := queue.PopNode()

		generation := ctx.EffectiveGeneration(idx)
		if generation > lastGeneration {
			return nil, errBadGenerationOrder
		}

		lastGeneration = generation
		if generation < minGeneration {
			break
		}

		flags := ctx.Marks(idx) & (markLeft | markRight | markStale)
		if flags == (markLeft | markRight) {
			if !ctx.HasAnyMarks(idx, markResult) {
				ctx.SetMarks(idx, markResult)
				results = append(results, idx)
			}

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

	out := results[:0]
	for _, idx := range results {
		if !ctx.HasAnyMarks(idx, markStale) {
			out = append(out, idx)
		}
	}

	return out, nil
}

func queueHasNonStale(ctx *Context, queue *PriorityQueue) bool {
	for _, idx := range queue.items {
		if !ctx.HasAnyMarks(idx, markStale) {
			return true
		}
	}

	return false
}
