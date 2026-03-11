package commitquery

// Marks returns the mark bits of one internal node.
func (query *Query) marks(idx nodeIndex) markBits {
	return query.nodes[idx].marks
}

// HasAnyMarks reports whether one internal node has any requested bit.
func (query *Query) hasAnyMarks(idx nodeIndex, bits markBits) bool {
	return query.nodes[idx].marks&bits != 0
}

// HasAllMarks reports whether one internal node already has all requested bits.
func (query *Query) hasAllMarks(idx nodeIndex, bits markBits) bool {
	return query.nodes[idx].marks&bits == bits
}

// SetMarks ORs one set of mark bits into one internal node.
func (query *Query) setMarks(idx nodeIndex, bits markBits) {
	newBits := bits &^ query.nodes[idx].marks
	if newBits == 0 {
		return
	}

	query.trackTouched(idx)
	query.nodes[idx].marks |= bits
}

// ClearMarks removes one set of mark bits from one internal node.
func (query *Query) clearMarks(idx nodeIndex, bits markBits) {
	if query.nodes[idx].marks&bits == 0 {
		return
	}

	query.trackTouched(idx)
	query.nodes[idx].marks &^= bits
}

// BeginMarkPhase starts one tracked mark-mutation phase.
func (query *Query) beginMarkPhase() {
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

// ClearTouchedMarks clears the provided bits from all nodes touched in the
// current mark phase.
func (query *Query) clearTouchedMarks(bits markBits) {
	for _, idx := range query.touched {
		query.nodes[idx].marks &^= bits
	}
}

func (query *Query) trackTouched(idx nodeIndex) {
	if query.nodes[idx].touchedPhase == query.markPhase {
		return
	}

	query.nodes[idx].touchedPhase = query.markPhase
	query.touched = append(query.touched, idx)
}
