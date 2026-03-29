package commitquery

// resolveParent resolves one parent descriptor to one internal node.
func (query *query) resolveParent(parent parentRef) (nodeIndex, error) {
	if parent.HasGraphPos {
		return query.resolveGraphPos(parent.GraphPos)
	}

	return query.resolveOID(parent.ID)
}
