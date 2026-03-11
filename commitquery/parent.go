package commitquery

import (
	commitgraphread "codeberg.org/lindenii/furgit/commitgraph/read"
	"codeberg.org/lindenii/furgit/objectid"
)

// parentRef references one commit parent.
type parentRef struct {
	ID          objectid.ObjectID
	GraphPos    commitgraphread.Position
	HasGraphPos bool
}

// Parents returns resolved parent node indices for one internal node.
func (query *Query) parents(idx nodeIndex) []nodeIndex {
	return query.nodes[idx].parents
}

// resolveParent resolves one parent descriptor to one internal node.
func (query *Query) resolveParent(parent parentRef) (nodeIndex, error) {
	if parent.HasGraphPos {
		return query.resolveGraphPos(parent.GraphPos)
	}

	return query.resolveOID(parent.ID)
}
