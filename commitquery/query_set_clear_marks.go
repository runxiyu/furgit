package commitquery

// setMarks ORs one set of mark bits into one internal node.
func (query *query) setMarks(idx nodeIndex, bits markBits) {
	newBits := bits &^ query.nodes[idx].marks
	if newBits == 0 {
		return
	}

	query.trackTouched(idx)
	query.nodes[idx].marks |= bits
}

// clearMarks removes one set of mark bits from one internal node.
func (query *query) clearMarks(idx nodeIndex, bits markBits) {
	if query.nodes[idx].marks&bits == 0 {
		return
	}

	query.trackTouched(idx)
	query.nodes[idx].marks &^= bits
}
