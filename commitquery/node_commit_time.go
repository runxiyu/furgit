package commitquery

// commitTime returns one node's commit time.
func (query *query) commitTime(idx nodeIndex) int64 {
	return query.nodes[idx].commitTime
}
