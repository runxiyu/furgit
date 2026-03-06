package sideband64k

// Band identifies the sideband stream within a pkt-line data frame.
type Band uint8

const (
	// BandData carries primary payload bytes.
	BandData Band = 1
	// BandProgress carries progress or informational messages.
	BandProgress Band = 2
	// BandError carries fatal error messages.
	BandError Band = 3
)
