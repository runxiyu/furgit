package commitquery

// markBits stores one set of traversal marks on one node.
type markBits uint8

// markLeft, markRight, markStale, and markResult track traversal state.
const (
	markLeft markBits = 1 << iota
	markRight
	markStale
	markResult
)

// allMarks is the union of all defined mark bits.
const (
	allMarks = markLeft | markRight | markStale | markResult
)
