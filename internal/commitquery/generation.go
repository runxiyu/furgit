package commitquery

import (
	"math"

	"codeberg.org/lindenii/furgit/objectid"
)

// EffectiveGeneration returns one node's generation value.
func (ctx *Context) EffectiveGeneration(idx NodeIndex) uint64 {
	if !ctx.nodes[idx].hasGeneration {
		return generationInfinity
	}

	return ctx.nodes[idx].generation
}

const (
	generationInfinity = uint64(math.MaxUint64)
)

func compareByGeneration(ctx *Context) func(NodeIndex, NodeIndex) int {
	return func(left, right NodeIndex) int {
		leftGeneration := ctx.EffectiveGeneration(left)
		rightGeneration := ctx.EffectiveGeneration(right)

		switch {
		case leftGeneration < rightGeneration:
			return -1
		case leftGeneration > rightGeneration:
			return 1
		}

		switch {
		case ctx.nodes[left].commitTime < ctx.nodes[right].commitTime:
			return -1
		case ctx.nodes[left].commitTime > ctx.nodes[right].commitTime:
			return 1
		}

		return objectid.Compare(ctx.nodes[left].id, ctx.nodes[right].id)
	}
}
