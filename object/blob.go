package object

import "codeberg.org/lindenii/furgit/objecttype"

// Blob represents a Git blob object.
type Blob struct {
	Data []byte
}

// ObjectType returns TypeBlob.
func (blob *Blob) ObjectType() objecttype.Type {
	_ = blob

	return objecttype.TypeBlob
}
