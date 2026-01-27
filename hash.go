package furgit

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
)

// maxHashSize MUST be >= the largest supported algorithm size.
const maxHashSize = sha256.Size

// hashAlgorithm identifies the hash algorithm used for Git object IDs.
type hashAlgorithm uint8

const (
	hashAlgoUnknown hashAlgorithm = iota
	hashAlgoSHA1
	hashAlgoSHA256
)

type hashAlgorithmDetails struct {
	name string
	size int
	sum  func([]byte) Hash
}

var hashAlgorithmTable = [...]hashAlgorithmDetails{
	hashAlgoUnknown: {},
	hashAlgoSHA1: {
		name: "sha1",
		size: sha1.Size,
		sum: func(data []byte) Hash {
			sum := sha1.Sum(data)
			var h Hash
			copy(h.data[:], sum[:])
			h.algo = hashAlgoSHA1
			return h
		},
	},
	hashAlgoSHA256: {
		name: "sha256",
		size: sha256.Size,
		sum: func(data []byte) Hash {
			sum := sha256.Sum256(data)
			var h Hash
			copy(h.data[:], sum[:])
			h.algo = hashAlgoSHA256
			return h
		},
	},
}

func (algo hashAlgorithm) info() hashAlgorithmDetails {
	return hashAlgorithmTable[algo]
}

// Size returns the hash size in bytes.
func (algo hashAlgorithm) Size() int {
	return algo.info().size
}

// String returns the canonical name of the hash algorithm.
func (algo hashAlgorithm) String() string {
	inf := algo.info()
	if inf.name == "" {
		return "unknown"
	}
	return inf.name
}

func (algo hashAlgorithm) HexLen() int {
	return algo.Size() * 2
}

func (algo hashAlgorithm) Sum(data []byte) Hash {
	return algo.info().sum(data)
}

// Hash represents a Git object ID.
type Hash struct {
	algo hashAlgorithm
	data [maxHashSize]byte
}

// String returns a hexadecimal string representation of the hash.
func (hash Hash) String() string {
	size := hash.algo.Size()
	if size == 0 {
		return ""
	}
	return hex.EncodeToString(hash.data[:size])
}

// Bytes returns a copy of the hash's bytes.
func (hash Hash) Bytes() []byte {
	size := hash.algo.Size()
	if size == 0 {
		return nil
	}
	return append([]byte(nil), hash.data[:size]...)
}

// Size returns the hash size.
func (hash Hash) Size() int {
	return hash.algo.Size()
}

var algoByName = map[string]hashAlgorithm{}

func init() {
	for algo, info := range hashAlgorithmTable {
		if info.name == "" {
			continue
		}
		algoByName[info.name] = hashAlgorithm(algo)
	}
}

func parseHashAlgorithm(s string) (hashAlgorithm, bool) {
	algo, ok := algoByName[s]
	return algo, ok
}
