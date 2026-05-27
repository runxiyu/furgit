package commitquery

import objectid "lindenii.org/go/furgit/object/id"

// IsAncestor reports whether ancestor is reachable from descendant through
// commit parent edges.
//
// Both inputs are peeled through annotated tags before commit traversal.
func (query *query) IsAncestor(ancestor, descendant objectid.ObjectID) (bool, error) {
	ancestorIdx, err := query.resolveCommitish(ancestor)
	if err != nil {
		return false, err
	}

	descendantIdx, err := query.resolveCommitish(descendant)
	if err != nil {
		return false, err
	}

	return query.isAncestor(ancestorIdx, descendantIdx)
}

// isAncestor answers one ancestry query between two resolved internal nodes.
func (query *query) isAncestor(ancestor, descendant nodeIndex) (bool, error) {
	if ancestor == descendant {
		return true, nil
	}

	ancestorGeneration := query.effectiveGeneration(ancestor)
	descendantGeneration := query.effectiveGeneration(descendant)

	if ancestorGeneration != generationInfinity &&
		descendantGeneration != generationInfinity &&
		ancestorGeneration > descendantGeneration {
		return false, nil
	}

	minGeneration := uint64(0)
	if ancestorGeneration != generationInfinity {
		minGeneration = ancestorGeneration
	}

	err := query.paintDownToCommon(ancestor, []nodeIndex{descendant}, minGeneration)
	if err != nil {
		return false, err
	}

	return query.hasAnyMarks(ancestor, markRight), nil
}
