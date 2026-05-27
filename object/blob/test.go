package blob

import objecttype "lindenii.org/go/furgit/object/type"

// ObjectType returns TypeBlob.
func (blob *Blob) ObjectType() objecttype.Type {
	_ = blob

	return objecttype.TypeBlob
}
