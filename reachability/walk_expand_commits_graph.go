package reachability

import (
	"errors"

	commitgraphread "codeberg.org/lindenii/furgit/format/commitgraph/read"
	objectid "codeberg.org/lindenii/furgit/object/id"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

func (walk *Walk) expandCommitsFromGraph(id objectid.ObjectID) ([]walkItem, bool, error) {
	pos, err := walk.reachability.graph.Lookup(id)
	if err != nil {
		if _, ok := errors.AsType[*commitgraphread.NotFoundError](err); ok {
			return nil, false, nil
		}

		return nil, true, err
	}

	commit, err := walk.reachability.graph.CommitAt(pos)
	if err != nil {
		return nil, true, err
	}

	next := make([]walkItem, 0, 2+len(commit.ExtraParents))

	if commit.Parent1.Valid {
		parentOID, err := walk.reachability.graph.OIDAt(commit.Parent1.Pos)
		if err != nil {
			return nil, true, err
		}

		next = append(next, walkItem{id: parentOID, want: objecttype.TypeInvalid})
	}

	if commit.Parent2.Valid {
		parentOID, err := walk.reachability.graph.OIDAt(commit.Parent2.Pos)
		if err != nil {
			return nil, true, err
		}

		next = append(next, walkItem{id: parentOID, want: objecttype.TypeInvalid})
	}

	for _, parentPos := range commit.ExtraParents {
		parentOID, err := walk.reachability.graph.OIDAt(parentPos)
		if err != nil {
			return nil, true, err
		}

		next = append(next, walkItem{id: parentOID, want: objecttype.TypeInvalid})
	}

	return next, true, nil
}
