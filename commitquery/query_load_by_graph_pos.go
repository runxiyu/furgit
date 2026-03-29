package commitquery

// loadByGraphPos populates one node from a commit-graph position.
func (query *query) loadByGraphPos(idx nodeIndex) error {
	pos := query.nodes[idx].graphPos

	return query.loadCommitAtGraphPos(idx, pos)
}
