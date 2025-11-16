package furgit

// ExtraHeader represents an extra header in a Git object.
type ExtraHeader struct {
	// Key represents the header key.
	Key string
	// Value represents the header value.
	Value []byte
}
