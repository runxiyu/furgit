package commitquery

import (
	commitgraphread "codeberg.org/lindenii/furgit/commitgraph/read"
	"codeberg.org/lindenii/furgit/objectid"
)

// Parent references one commit parent.
type Parent struct {
	ID          objectid.ObjectID
	GraphPos    commitgraphread.Position
	HasGraphPos bool
}

// Parents returns resolved parent node indices for one internal node.
func (ctx *Context) Parents(idx NodeIndex) []NodeIndex {
	return ctx.nodes[idx].parents
}

// resolveParent resolves one parent descriptor to one internal node.
func (ctx *Context) resolveParent(parent Parent) (NodeIndex, error) {
	if parent.HasGraphPos {
		return ctx.ResolveGraphPos(parent.GraphPos)
	}

	return ctx.ResolveOID(parent.ID)
}
