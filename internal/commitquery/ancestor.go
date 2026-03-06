package commitquery

// IsAncestor reports whether ancestor is reachable from descendant through
// commit parent edges.
func IsAncestor(ctx *Context, ancestor, descendant NodeIndex) (bool, error) {
	if ancestor == descendant {
		return true, nil
	}

	ancestorGeneration := ctx.EffectiveGeneration(ancestor)
	descendantGeneration := ctx.EffectiveGeneration(descendant)

	if ancestorGeneration != generationInfinity &&
		descendantGeneration != generationInfinity &&
		ancestorGeneration > descendantGeneration {
		return false, nil
	}

	minGeneration := uint64(0)
	if ancestorGeneration != generationInfinity {
		minGeneration = ancestorGeneration
	}

	_, err := paintDownToCommon(ctx, ancestor, []NodeIndex{descendant}, minGeneration)
	if err != nil {
		return false, err
	}

	return ctx.HasAnyMarks(ancestor, markRight), nil
}
