package id

import (
	"bytes"
	"encoding/hex"
)

// Algorithm returns the object ID's hash algorithm.
func (id ObjectID) Algorithm() Algorithm {
	return id.algo
}

// Bytes returns a copy of the object ID bytes.
func (id ObjectID) Bytes() []byte {
	size := id.Algorithm().Size()

	return append([]byte(nil), id.data[:size]...)
}

// RawBytes returns a direct byte slice view of the object ID bytes.
//
// Prefer [ObjectID.Bytes] except for when it is a performance bottleneck.
//
// Labels: Mut-No.
func (id *ObjectID) RawBytes() []byte {
	size := id.Algorithm().Size()

	return id.data[:size:size]
}

// Compare lexicographically compares two object IDs
// by their canonical byte representation.
func (id ObjectID) Compare(other ObjectID) int {
	return bytes.Compare(id.RawBytes(), other.RawBytes())
}

// String returns the canonical hex representation.
func (id ObjectID) String() string {
	size := id.Algorithm().Size()

	return hex.EncodeToString(id.data[:size])
}
