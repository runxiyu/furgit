package commitquery

import (
	"slices"
)

// removeRedundant removes redundant merge-base candidates.
func removeRedundant(ctx *Context, candidates []NodeIndex) ([]NodeIndex, error) {
	for _, idx := range candidates {
		if ctx.EffectiveGeneration(idx) != generationInfinity {
			return removeRedundantWithGen(ctx, candidates), nil
		}
	}

	return removeRedundantNoGen(ctx, candidates)
}

func removeRedundantNoGen(ctx *Context, candidates []NodeIndex) ([]NodeIndex, error) {
	redundant := make([]bool, len(candidates))
	work := make([]NodeIndex, 0, len(candidates)-1)
	filledIndex := make([]int, 0, len(candidates)-1)

	for i, candidate := range candidates {
		if redundant[i] {
			continue
		}

		work = work[:0]
		filledIndex = filledIndex[:0]

		minGeneration := ctx.EffectiveGeneration(candidate)

		for j, other := range candidates {
			if i == j || redundant[j] {
				continue
			}

			work = append(work, other)
			filledIndex = append(filledIndex, j)

			otherGeneration := ctx.EffectiveGeneration(other)
			if otherGeneration < minGeneration {
				minGeneration = otherGeneration
			}
		}

		err := paintDownToCommon(ctx, candidate, work, minGeneration)
		if err != nil {
			return nil, err
		}

		if ctx.HasAnyMarks(candidate, markRight) {
			redundant[i] = true
		}

		for j, other := range work {
			if ctx.HasAnyMarks(other, markLeft) {
				redundant[filledIndex[j]] = true
			}
		}

		ctx.ClearTouchedMarks(allMarks)
	}

	out := make([]NodeIndex, 0, len(candidates))
	for i, idx := range candidates {
		if !redundant[i] {
			out = append(out, idx)
		}
	}

	return out, nil
}

func removeRedundantWithGen(ctx *Context, candidates []NodeIndex) []NodeIndex {
	sorted := append([]NodeIndex(nil), candidates...)
	slices.SortFunc(sorted, compareByGeneration(ctx))

	minGeneration := ctx.EffectiveGeneration(sorted[0])
	minGenPos := 0
	countStillIndependent := len(candidates)

	ctx.BeginMarkPhase()

	walkStart := make([]NodeIndex, 0, len(candidates)*2)

	for _, idx := range candidates {
		ctx.SetMarks(idx, markResult)

		for _, parent := range ctx.Parents(idx) {
			if ctx.HasAnyMarks(parent, markStale) {
				continue
			}

			ctx.SetMarks(parent, markStale)
			walkStart = append(walkStart, parent)
		}
	}

	slices.SortFunc(walkStart, compareByGeneration(ctx))

	for _, idx := range walkStart {
		ctx.ClearMarks(idx, markStale)
	}

	for i := len(walkStart) - 1; i >= 0 && countStillIndependent > 1; i-- {
		stack := []NodeIndex{walkStart[i]}
		ctx.SetMarks(walkStart[i], markStale)

		for len(stack) > 0 {
			top := stack[len(stack)-1]

			if ctx.HasAnyMarks(top, markResult) {
				ctx.ClearMarks(top, markResult)

				countStillIndependent--
				if countStillIndependent <= 1 {
					break
				}

				if top == sorted[minGenPos] {
					for minGenPos < len(sorted)-1 && ctx.HasAnyMarks(sorted[minGenPos], markStale) {
						minGenPos++
					}

					minGeneration = ctx.EffectiveGeneration(sorted[minGenPos])
				}
			}

			if ctx.EffectiveGeneration(top) < minGeneration {
				stack = stack[:len(stack)-1]

				continue
			}

			pushed := false

			for _, parent := range ctx.Parents(top) {
				if ctx.HasAnyMarks(parent, markStale) {
					continue
				}

				ctx.SetMarks(parent, markStale)
				stack = append(stack, parent)
				pushed = true

				break
			}

			if !pushed {
				stack = stack[:len(stack)-1]
			}
		}
	}

	out := make([]NodeIndex, 0, len(candidates))
	for _, idx := range candidates {
		if !ctx.HasAnyMarks(idx, markStale) {
			out = append(out, idx)
		}
	}

	ctx.ClearTouchedMarks(markStale | markResult)

	return out
}
