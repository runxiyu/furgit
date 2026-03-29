package commitquery

// marks returns the mark bits of one internal node.
func (query *query) marks(idx nodeIndex) markBits {
	return query.nodes[idx].marks
}
