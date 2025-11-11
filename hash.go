package furgit

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
)

// To change the hash algorithm you probably only need to change these two lines...

const HashSize = sha1.Size

var newHash = sha1.Sum

// Hash represents a Git object identifier.
type Hash [HashSize]byte

// ParseHash converts a hex string into an Hash.
func ParseHash(s string) (Hash, error) {
	var id Hash
	if len(s) != HashSize*2 {
		return id, fmt.Errorf("furgit: invalid hash length %d", len(s))
	}
	data, err := hex.DecodeString(s)
	if err != nil {
		return id, fmt.Errorf("furgit: decode hash: %w", err)
	}
	copy(id[:], data)
	return id, nil
}

// String renders the ID as hex.
func (id Hash) String() string {
	return hex.EncodeToString(id[:])
}

// Bytes returns a mutable copy of the underlying bytes.
func (id Hash) Bytes() []byte {
	return append([]byte(nil), id[:]...)
}
