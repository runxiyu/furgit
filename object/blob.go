package object

import objecttype "codeberg.org/lindenii/furgit/object/type"

// Blob represents a Git blob object.
//
// This Blob object is fully materialized in memory.
// Consider using objectstore/Store.ReadReaderContent,
// or appropriate streaming write APIs.
type Blob struct {
	Data []byte
}

// ObjectType returns TypeBlob.
func (blob *Blob) ObjectType() objecttype.Type {
	_ = blob

	return objecttype.TypeBlob
}
