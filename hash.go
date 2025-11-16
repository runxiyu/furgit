package furgit

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
)

const maxHashSize = 32

// Hash represents a Git object identifier.
type Hash struct {
	data [maxHashSize]byte
	size int
}

// hashFunc is a function that computes a hash from input data.
type hashFunc func([]byte) Hash

// hashFuncs maps hash size to hash function.
var hashFuncs = map[int]hashFunc{
	sha1.Size: func(data []byte) Hash {
		sum := sha1.Sum(data)
		var h Hash
		copy(h.data[:], sum[:])
		h.size = sha1.Size
		return h
	},
	sha256.Size: func(data []byte) Hash {
		sum := sha256.Sum256(data)
		var h Hash
		copy(h.data[:], sum[:])
		h.size = sha256.Size
		return h
	},
}

// String returns the ID as hex using its internal size.
func (hash Hash) String() string {
	return hex.EncodeToString(hash.data[:hash.size])
}

// Bytes returns a mutable copy of the underlying bytes using its internal size.
func (hash Hash) Bytes() []byte {
	return append([]byte(nil), hash.data[:hash.size]...)
}

// Size returns the hash size.
func (hash Hash) Size() int {
	return hash.size
}
