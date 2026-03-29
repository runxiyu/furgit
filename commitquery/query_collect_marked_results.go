package commitquery

// collectMarkedResults returns touched nodes marked as non-stale results.
func (query *query) collectMarkedResults() []nodeIndex {
	out := make([]nodeIndex, 0, 4)

	for _, idx := range query.touched {
		if !query.hasAnyMarks(idx, markResult) {
			continue
		}

		if query.hasAnyMarks(idx, markStale) {
			continue
		}

		out = append(out, idx)
	}

	return out
}
