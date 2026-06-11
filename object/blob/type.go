package blob

import "lindenii.org/go/furgit/object/typ"

// ObjectType returns TypeBlob.
func (blob *Blob) ObjectType() typ.Type {
	_ = blob

	return typ.Blob
}
