package furgit

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
)

// maxHashSize MUST be equal to (or larger than) the size of the
// largest hash supported in hashFuncs.
const maxHashSize = sha256.Size

// hashAlgorithm identifies the hash algorithm used for Git object IDs.
type hashAlgorithm uint8

// hashFuncs maps hash algorithm to hash function.
var hashFuncs = map[hashAlgorithm]hashFunc{
	hashAlgoSHA1: func(data []byte) Hash {
		sum := sha1.Sum(data)
		var h Hash
		copy(h.data[:], sum[:])
		h.algo = hashAlgoSHA1
		return h
	},
	hashAlgoSHA256: func(data []byte) Hash {
		sum := sha256.Sum256(data)
		var h Hash
		copy(h.data[:], sum[:])
		h.algo = hashAlgoSHA256
		return h
	},
}

const (
	hashAlgoUnknown hashAlgorithm = iota
	hashAlgoSHA1
	hashAlgoSHA256
)

// size returns the hash size in bytes.
func (algo hashAlgorithm) size() int {
	switch algo {
	case hashAlgoSHA1:
		return sha1.Size
	case hashAlgoSHA256:
		return sha256.Size
	default:
		return 0
	}
}

// String returns the canonical name of the hash algorithm.
func (algo hashAlgorithm) String() string {
	switch algo {
	case hashAlgoSHA1:
		return "sha1"
	case hashAlgoSHA256:
		return "sha256"
	default:
		return "unknown"
	}
}

// Hash represents a Git object ID.
type Hash struct {
	algo hashAlgorithm
	data [maxHashSize]byte
}

// hashFunc is a function that computes a hash from input data.
type hashFunc func([]byte) Hash

// String returns a hexadecimal string representation of the hash.
func (hash Hash) String() string {
	size := hash.algo.size()
	if size == 0 {
		return ""
	}
	return hex.EncodeToString(hash.data[:size])
}

// Bytes returns a copy of the hash's bytes.
func (hash Hash) Bytes() []byte {
	size := hash.algo.size()
	if size == 0 {
		return nil
	}
	return append([]byte(nil), hash.data[:size]...)
}

// Size returns the hash size.
func (hash Hash) Size() int {
	return hash.algo.size()
}
