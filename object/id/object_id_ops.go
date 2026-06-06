package id

import (
	"bytes"
	"encoding/hex"
)

// ObjectFormat returns the object ID's object format (hash algorithm).
func (id ObjectID) ObjectFormat() ObjectFormat {
	return id.objectFormat
}

// Bytes returns a copy of the object ID bytes.
//
// Labels: Life-Independent.
func (id ObjectID) Bytes() []byte {
	size := id.ObjectFormat().Size()

	return append([]byte(nil), id.data[:size]...)
}

// RawBytes returns a direct byte slice view of the object ID bytes.
//
// Prefer [ObjectID.Bytes] except for when it is a performance bottleneck.
//
// Labels: Life-Parent, Mut-No.
func (id *ObjectID) RawBytes() []byte {
	size := id.ObjectFormat().Size()

	return id.data[:size:size]
}

// Compare lexicographically compares two object IDs
// by their canonical byte representation.
func (id ObjectID) Compare(other ObjectID) int {
	return bytes.Compare(id.RawBytes(), other.RawBytes())
}

// String returns the canonical hex representation.
func (id ObjectID) String() string {
	size := id.ObjectFormat().Size()

	return hex.EncodeToString(id.data[:size])
}
