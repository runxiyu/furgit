package blob

import objecttype "codeberg.org/lindenii/furgit/object/type"

// ObjectType returns TypeBlob.
func (blob *Blob) ObjectType() objecttype.Type {
	_ = blob

	return objecttype.TypeBlob
}
