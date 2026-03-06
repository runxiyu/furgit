package reachability

import (
	"errors"

	commitgraphread "codeberg.org/lindenii/furgit/format/commitgraph/read"
	"codeberg.org/lindenii/furgit/objectid"
)

// IsAncestor reports whether ancestor is reachable from descendant via commit
// parent edges.
//
// Both inputs are peeled through annotated tags before commit traversal.
func (r *Reachability) IsAncestor(ancestor, descendant objectid.ObjectID) (bool, error) {
	ancestorCommit, err := r.peelRootToDomain(ancestor, DomainCommits)
	if err != nil {
		return false, err
	}

	descendantCommit, err := r.peelRootToDomain(descendant, DomainCommits)
	if err != nil {
		return false, err
	}

	if ancestorCommit == descendantCommit {
		return true, nil
	}

	graphResult, graphUsed, err := r.isAncestorGraph(ancestorCommit, descendantCommit)
	if err != nil {
		return false, err
	}

	if graphUsed {
		return graphResult, nil
	}

	walk := r.Walk(DomainCommits, nil, map[objectid.ObjectID]struct{}{descendantCommit: {}})
	for id := range walk.Seq() {
		if id == ancestorCommit {
			return true, nil
		}
	}

	err = walk.Err()
	if err != nil {
		return false, err
	}

	return false, nil
}

func (r *Reachability) isAncestorGraph(ancestor, descendant objectid.ObjectID) (bool, bool, error) {
	if r.graph == nil {
		return false, false, nil
	}

	ancestorPos, err := r.graph.Lookup(ancestor)
	if err != nil {
		var notFound *commitgraphread.NotFoundError
		if errors.As(err, &notFound) {
			return false, false, nil
		}

		return false, true, err
	}

	descendantPos, err := r.graph.Lookup(descendant)
	if err != nil {
		var notFound *commitgraphread.NotFoundError
		if errors.As(err, &notFound) {
			return false, false, nil
		}

		return false, true, err
	}

	ancestorCommit, err := r.graph.CommitAt(ancestorPos)
	if err != nil {
		return false, true, err
	}

	ancestorGeneration := ancestorCommit.GenerationV2
	stack := []commitgraphread.Position{descendantPos}
	visited := make(map[commitgraphread.Position]struct{}, 64)

	for len(stack) > 0 {
		pos := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if _, ok := visited[pos]; ok {
			continue
		}

		visited[pos] = struct{}{}

		if pos == ancestorPos {
			return true, true, nil
		}

		commit, err := r.graph.CommitAt(pos)
		if err != nil {
			return false, true, err
		}

		if commit.GenerationV2 < ancestorGeneration {
			continue
		}

		if commit.Parent1.Valid {
			stack = append(stack, commit.Parent1.Pos)
		}

		if commit.Parent2.Valid {
			stack = append(stack, commit.Parent2.Pos)
		}

		stack = append(stack, commit.ExtraParents...)
	}

	return false, true, nil
}
