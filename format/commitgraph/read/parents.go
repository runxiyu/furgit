package read

import "codeberg.org/lindenii/furgit/format/commitgraph"

// ParentRef references one parent position.
type ParentRef struct {
	Valid bool
	Pos   Position
}

func (reader *Reader) decodeParents(layer *layer, p1, p2 uint32) (ParentRef, ParentRef, []Position, error) {
	parent1, err := reader.decodeSingleParent(p1)
	if err != nil {
		return ParentRef{}, ParentRef{}, nil, err
	}

	if p2 == commitgraph.ParentNone {
		return parent1, ParentRef{}, nil, nil
	}

	if p2&commitgraph.ParentExtraMask == 0 {
		parent2, err := reader.decodeSingleParent(p2)
		if err != nil {
			return ParentRef{}, ParentRef{}, nil, err
		}

		return parent1, parent2, nil, nil
	}

	edgeStart := p2 & commitgraph.ParentLastMask

	parents, err := reader.decodeExtraEdgeList(layer, edgeStart)
	if err != nil {
		return ParentRef{}, ParentRef{}, nil, err
	}

	if len(parents) == 0 {
		return ParentRef{}, ParentRef{}, nil, &MalformedError{Path: layer.path, Reason: "empty EDGE list"}
	}

	parent2 := ParentRef{Valid: true, Pos: parents[0]}
	if len(parents) == 1 {
		return parent1, parent2, nil, nil
	}

	return parent1, parent2, parents[1:], nil
}

func (reader *Reader) decodeSingleParent(raw uint32) (ParentRef, error) {
	if raw == commitgraph.ParentNone {
		return ParentRef{}, nil
	}

	if raw&commitgraph.ParentExtraMask != 0 {
		return ParentRef{}, &MalformedError{
			Path:   "commit-graph",
			Reason: "unexpected EDGE marker in single-parent slot",
		}
	}

	pos, err := reader.globalToPosition(raw)
	if err != nil {
		return ParentRef{}, err
	}

	return ParentRef{Valid: true, Pos: pos}, nil
}
