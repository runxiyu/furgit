package commitquery

// acquire removes one worker from the idle pool or allocates one new worker.
func (queries *Queries) acquire() *query {
	queries.mu.Lock()
	defer queries.mu.Unlock()

	count := len(queries.idle)
	if count == 0 {
		return newQuery(queries.fetcher, queries.graph)
	}

	q := queries.idle[count-1]
	queries.idle = queries.idle[:count-1]

	return q
}
