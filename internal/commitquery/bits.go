package commitquery

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
