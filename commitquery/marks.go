package commitquery

// Marks returns the mark bits of one internal node.
func (query *query) marks(idx nodeIndex) markBits {
	return query.nodes[idx].marks
}

// HasAnyMarks reports whether one internal node has any requested bit.
func (query *query) hasAnyMarks(idx nodeIndex, bits markBits) bool {
	return query.nodes[idx].marks&bits != 0
}

// HasAllMarks reports whether one internal node already has all requested bits.
func (query *query) hasAllMarks(idx nodeIndex, bits markBits) bool {
	return query.nodes[idx].marks&bits == bits
}

// SetMarks ORs one set of mark bits into one internal node.
func (query *query) setMarks(idx nodeIndex, bits markBits) {
	newBits := bits &^ query.nodes[idx].marks
	if newBits == 0 {
		return
	}

	query.trackTouched(idx)
	query.nodes[idx].marks |= bits
}

// ClearMarks removes one set of mark bits from one internal node.
func (query *query) clearMarks(idx nodeIndex, bits markBits) {
	if query.nodes[idx].marks&bits == 0 {
		return
	}

	query.trackTouched(idx)
	query.nodes[idx].marks &^= bits
}

// BeginMarkPhase starts one tracked mark-mutation phase.
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

// ClearTouchedMarks clears the provided bits from all nodes touched in the
// current mark phase.
func (query *query) clearTouchedMarks(bits markBits) {
	for _, idx := range query.touched {
		query.nodes[idx].marks &^= bits
	}
}

func (query *query) trackTouched(idx nodeIndex) {
	if query.nodes[idx].touchedPhase == query.markPhase {
		return
	}

	query.nodes[idx].touchedPhase = query.markPhase
	query.touched = append(query.touched, idx)
}

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

type markBits uint8

const (
	markLeft markBits = 1 << iota
	markRight
	markStale
	markResult
)

const (
	allMarks = markLeft | markRight | markStale | markResult
)
