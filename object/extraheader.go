package object

// ExtraHeader represents an extra header in a Git object.
type ExtraHeader struct {
	Key   string
	Value []byte
}
