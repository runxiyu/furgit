package objectid

//#nosec gosec

// Algorithm identifies the hash algorithm used for Git object IDs.
type Algorithm uint8

const (
	AlgorithmUnknown Algorithm = iota
	AlgorithmSHA1
	AlgorithmSHA256
)
