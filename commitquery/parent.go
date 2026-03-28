package commitquery

import (
	commitgraphread "codeberg.org/lindenii/furgit/format/commitgraph/read"
	objectid "codeberg.org/lindenii/furgit/object/id"
)

// parentRef references one commit parent.
type parentRef struct {
	ID          objectid.ObjectID
	GraphPos    commitgraphread.Position
	HasGraphPos bool
}

// Parents returns resolved parent node indices for one internal node.
func (query *query) parents(idx nodeIndex) []nodeIndex {
	return query.nodes[idx].parents
}

// resolveParent resolves one parent descriptor to one internal node.
func (query *query) resolveParent(parent parentRef) (nodeIndex, error) {
	if parent.HasGraphPos {
		return query.resolveGraphPos(parent.GraphPos)
	}

	return query.resolveOID(parent.ID)
}
