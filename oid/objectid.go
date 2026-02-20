// Package oid provides object ID and algorithm primitives for Git objects.
package oid

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
)

var (
	// ErrInvalidAlgorithm indicates an unsupported object ID algorithm.
	ErrInvalidAlgorithm = errors.New("oid: invalid algorithm")
	// ErrInvalidObjectID indicates malformed object ID data.
	ErrInvalidObjectID = errors.New("oid: invalid object id")
)

// maxObjectIDSize MUST be >= the largest supported algorithm size.
const maxObjectIDSize = sha256.Size

// Algorithm identifies the hash algorithm used for Git object IDs.
type Algorithm uint8

const (
	AlgorithmUnknown Algorithm = iota
	AlgorithmSHA1
	AlgorithmSHA256
)

type algorithmDetails struct {
	name string
	size int
	sum  func([]byte) ObjectID
	new  func() hash.Hash
}

var algorithmTable = [...]algorithmDetails{
	AlgorithmUnknown: {},
	AlgorithmSHA1: {
		name: "sha1",
		size: sha1.Size,
		sum: func(data []byte) ObjectID {
			sum := sha1.Sum(data)
			var id ObjectID
			copy(id.data[:], sum[:])
			id.algo = AlgorithmSHA1
			return id
		},
		new: func() hash.Hash {
			return sha1.New()
		},
	},
	AlgorithmSHA256: {
		name: "sha256",
		size: sha256.Size,
		sum: func(data []byte) ObjectID {
			sum := sha256.Sum256(data)
			var id ObjectID
			copy(id.data[:], sum[:])
			id.algo = AlgorithmSHA256
			return id
		},
		new: func() hash.Hash {
			return sha256.New()
		},
	},
}

var algorithmByName = map[string]Algorithm{}

func init() {
	for algo, info := range algorithmTable {
		if info.name == "" {
			continue
		}
		algorithmByName[info.name] = Algorithm(algo)
	}
}

func (algo Algorithm) info() algorithmDetails {
	return algorithmTable[algo]
}

// ParseAlgorithm parses a canonical algorithm name (e.g. "sha1", "sha256").
func ParseAlgorithm(s string) (Algorithm, bool) {
	algo, ok := algorithmByName[s]
	return algo, ok
}

// Size returns the hash size in bytes.
func (algo Algorithm) Size() int {
	return algo.info().size
}

// String returns the canonical algorithm name.
func (algo Algorithm) String() string {
	inf := algo.info()
	if inf.name == "" {
		return "unknown"
	}
	return inf.name
}

// HexLen returns the encoded hexadecimal length.
func (algo Algorithm) HexLen() int {
	return algo.Size() * 2
}

// Sum computes an object ID from raw data using the selected algorithm.
func (algo Algorithm) Sum(data []byte) ObjectID {
	return algo.info().sum(data)
}

// New returns a new hash.Hash for this algorithm.
func (algo Algorithm) New() (hash.Hash, error) {
	newFn := algo.info().new
	if newFn == nil {
		return nil, ErrInvalidAlgorithm
	}
	return newFn(), nil
}

// ObjectID represents a Git object ID.
type ObjectID struct {
	algo Algorithm
	data [maxObjectIDSize]byte
}

// Algorithm returns the object ID's hash algorithm.
func (id ObjectID) Algorithm() Algorithm {
	return id.algo
}

// Size returns the object ID size in bytes.
func (id ObjectID) Size() int {
	return id.algo.Size()
}

// String returns the canonical hex representation.
func (id ObjectID) String() string {
	size := id.Size()
	return hex.EncodeToString(id.data[:size])
}

// Bytes returns a copy of the object ID bytes.
func (id ObjectID) Bytes() []byte {
	size := id.Size()
	return append([]byte(nil), id.data[:size]...)
}

// ParseHex parses an object ID from hex for the specified algorithm.
func ParseHex(algo Algorithm, s string) (ObjectID, error) {
	var id ObjectID
	if algo.Size() == 0 {
		return id, ErrInvalidAlgorithm
	}
	if len(s)%2 != 0 {
		return id, fmt.Errorf("%w: odd hex length %d", ErrInvalidObjectID, len(s))
	}
	if len(s) != algo.HexLen() {
		return id, fmt.Errorf("%w: got %d chars, expected %d", ErrInvalidObjectID, len(s), algo.HexLen())
	}
	decoded, err := hex.DecodeString(s)
	if err != nil {
		return id, fmt.Errorf("%w: decode: %v", ErrInvalidObjectID, err)
	}
	copy(id.data[:], decoded)
	id.algo = algo
	return id, nil
}

// FromBytes builds an object ID from raw bytes for the specified algorithm.
func FromBytes(algo Algorithm, b []byte) (ObjectID, error) {
	var id ObjectID
	if algo.Size() == 0 {
		return id, ErrInvalidAlgorithm
	}
	if len(b) != algo.Size() {
		return id, fmt.Errorf("%w: got %d bytes, expected %d", ErrInvalidObjectID, len(b), algo.Size())
	}
	copy(id.data[:], b)
	id.algo = algo
	return id, nil
}
