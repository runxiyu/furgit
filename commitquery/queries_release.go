package commitquery

func (queries *Queries) release(q *query) {
	q.resetForReuse()

	queries.mu.Lock()
	defer queries.mu.Unlock()

	if len(queries.idle) >= queries.maxIdle {
		return
	}

	queries.idle = append(queries.idle, q)
}
