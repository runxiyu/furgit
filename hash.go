package furgit

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

const maxHashSize = 32

// Hash represents a Git object identifier.
type Hash [maxHashSize]byte

// hashFunc is a function that computes a hash from input data.
type hashFunc func([]byte) [maxHashSize]byte

// hashFuncs maps hash size to hash function.
var hashFuncs = map[int]hashFunc{
	sha1.Size: func(data []byte) [maxHashSize]byte {
		var result [maxHashSize]byte
		sum := sha1.Sum(data)
		copy(result[:], sum[:])
		return result
	},
	sha256.Size: func(data []byte) [maxHashSize]byte {
		var result [maxHashSize]byte
		sum := sha256.Sum256(data)
		copy(result[:], sum[:])
		return result
	},
}

// ParseHashWithSize converts a hex string into a Hash for a given hash size.
func ParseHashWithSize(s string, hashSize int) (Hash, error) {
	var id Hash
	if len(s) != hashSize*2 {
		return id, fmt.Errorf("furgit: invalid hash length %d", len(s))
	}
	data, err := hex.DecodeString(s)
	if err != nil {
		return id, fmt.Errorf("furgit: decode hash: %w", err)
	}
	copy(id[:], data)
	return id, nil
}

// StringWithSize returns the ID as hex for a given hash size.
func (id Hash) StringWithSize(hashSize int) string {
	return hex.EncodeToString(id[:hashSize])
}

// BytesWithSize returns a mutable copy of the underlying bytes for a given hash size.
func (id Hash) BytesWithSize(hashSize int) []byte {
	return append([]byte(nil), id[:hashSize]...)
}
