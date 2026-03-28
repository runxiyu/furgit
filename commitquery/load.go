package commitquery

// ensureLoaded completes one node's metadata load if it has not been loaded yet.
func (query *query) ensureLoaded(idx nodeIndex) error {
	if query.nodes[idx].loaded {
		return nil
	}

	if query.nodes[idx].hasGraphPos {
		return query.loadByGraphPos(idx)
	}

	return query.loadByOID(idx)
}
