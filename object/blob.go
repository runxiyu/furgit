package object

// Blob represents a Git blob object.
type Blob struct {
	Data []byte
}

// ObjectType returns TypeBlob.
func (blob *Blob) ObjectType() Type {
	_ = blob
	return TypeBlob
}
