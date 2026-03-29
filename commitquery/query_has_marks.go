package commitquery

// hasAnyMarks reports whether one internal node has any requested bit.
func (query *query) hasAnyMarks(idx nodeIndex, bits markBits) bool {
	return query.nodes[idx].marks&bits != 0
}

// hasAllMarks reports whether one internal node already has all requested bits.
func (query *query) hasAllMarks(idx nodeIndex, bits markBits) bool {
	return query.nodes[idx].marks&bits == bits
}
