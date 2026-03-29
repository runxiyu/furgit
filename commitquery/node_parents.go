package commitquery

// parents returns resolved parent node indices for one internal node.
func (query *query) parents(idx nodeIndex) []nodeIndex {
	return query.nodes[idx].parents
}
