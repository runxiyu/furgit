package commitquery

import "codeberg.org/lindenii/furgit/objectid"

// Compare compares two internal nodes using merge-base queue ordering.
func (ctx *Context) Compare(left, right NodeIndex) int {
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
