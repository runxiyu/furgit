package commitquery

// resetForReuse clears transient state before one worker returns to the pool.
func (query *query) resetForReuse() {
	for _, idx := range query.touched {
		query.nodes[idx].marks = 0
	}

	query.touched = query.touched[:0]
}
