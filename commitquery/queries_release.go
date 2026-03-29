package commitquery

// release resets one worker and returns it to the idle pool if there is room.
func (queries *Queries) release(q *query) {
	q.resetForReuse()

	queries.mu.Lock()
	defer queries.mu.Unlock()

	if len(queries.idle) >= queries.maxIdle {
		return
	}

	queries.idle = append(queries.idle, q)
}
