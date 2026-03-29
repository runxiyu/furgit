package commitquery

func (query *query) commitTime(idx nodeIndex) int64 {
	return query.nodes[idx].commitTime
}
