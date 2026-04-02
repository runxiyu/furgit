package id

// Algorithm identifies the hash algorithm used for Git object IDs.
type Algorithm uint8

const (
	// AlgorithmUnknown identifies an unknown object ID hash algorithm.
	AlgorithmUknown Algorithm = iota

	// AlgorithmSHA1 identifies the SHA-1 object ID hash algorithm.
	// This is the default for all versions of Git until Git 3.0.
	AlgorithmSHA1

	// AlgorithmSHA256 identifies the SHA-256 object ID hash algorithm.
	// This is the default for Git 3.0 and beyond.
	AlgorithmSHA256
)
