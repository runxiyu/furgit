package blob

import "codeberg.org/lindenii/furgit/object/typ"

// ObjectType returns TypeBlob.
func (blob *Blob) ObjectType() typ.Type {
	_ = blob

	return typ.TypeBlob
}
