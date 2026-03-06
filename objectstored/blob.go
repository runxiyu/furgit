package objectstored

import (
	"codeberg.org/lindenii/furgit/object"
	"codeberg.org/lindenii/furgit/objectid"
)

// StoredBlob is a parsed blob paired with its storage ID.
//
// This Blob object is fully materialized in memory.
// Consider using objectstore/Store.ReadReaderContent.
type StoredBlob struct {
	id   objectid.ObjectID
	blob *object.Blob
}

// NewStoredBlob creates one stored blob wrapper.
func NewStoredBlob(id objectid.ObjectID, blob *object.Blob) *StoredBlob {
	return &StoredBlob{id: id, blob: blob}
}

// ID returns the object ID this blob was loaded from.
func (stored *StoredBlob) ID() objectid.ObjectID {
	return stored.id
}

// Object returns the parsed blob as the generic object interface.
//
//nolint:ireturn
func (stored *StoredBlob) Object() object.Object {
	return stored.blob
}

// Blob returns the parsed blob value.
func (stored *StoredBlob) Blob() *object.Blob {
	return stored.blob
}
