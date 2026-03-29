package commitquery

// beginMarkPhase starts one tracked mark-mutation phase.
func (query *query) beginMarkPhase() {
	for _, idx := range query.touched {
		query.nodes[idx].marks = 0
	}

	query.markPhase++
	if query.markPhase == 0 {
		query.markPhase++
		for i := range query.nodes {
			query.nodes[i].touchedPhase = 0
		}
	}

	query.touched = query.touched[:0]
}

// clearTouchedMarks clears the provided bits from all nodes touched in the
// current mark phase.
func (query *query) clearTouchedMarks(bits markBits) {
	for _, idx := range query.touched {
		query.nodes[idx].marks &^= bits
	}
}

// trackTouched records one node in the current mark phase.
func (query *query) trackTouched(idx nodeIndex) {
	if query.nodes[idx].touchedPhase == query.markPhase {
		return
	}

	query.nodes[idx].touchedPhase = query.markPhase
	query.touched = append(query.touched, idx)
}
