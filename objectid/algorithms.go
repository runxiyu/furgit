package objectid

import (
	"crypto/sha1" //#nosec gosec
	"crypto/sha256"
	"hash"
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
	name       string
	size       int
	packHashID uint32
	sum        func([]byte) ObjectID
	new        func() hash.Hash
}

var algorithmTable = [...]algorithmDetails{
	AlgorithmUnknown: {},
	AlgorithmSHA1: {
		name:       "sha1",
		size:       sha1.Size,
		packHashID: 1,
		sum: func(data []byte) ObjectID {
			sum := sha1.Sum(data) //#nosec G401

			var id ObjectID
			copy(id.data[:], sum[:])
			id.algo = AlgorithmSHA1

			return id
		},
		new: sha1.New,
	},
	AlgorithmSHA256: {
		name:       "sha256",
		size:       sha256.Size,
		packHashID: 2,
		sum: func(data []byte) ObjectID {
			sum := sha256.Sum256(data)

			var id ObjectID
			copy(id.data[:], sum[:])
			id.algo = AlgorithmSHA256

			return id
		},
		new: sha256.New,
	},
}

var (
	algorithmByName     = map[string]Algorithm{}
	supportedAlgorithms []Algorithm
)

func init() { //nolint:gochecknoinits
	for algo := Algorithm(0); int(algo) < len(algorithmTable); algo++ {
		info := algorithmTable[algo]
		if info.name == "" {
			continue
		}

		algorithmByName[info.name] = algo
		supportedAlgorithms = append(supportedAlgorithms, algo)
	}
}

// SupportedAlgorithms returns all object ID algorithms supported by furgit.
// Do not mutate.
func SupportedAlgorithms() []Algorithm {
	return supportedAlgorithms
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

// PackHashID returns the Git pack/rev hash-id encoding for this algorithm.
//
// Unknown algorithms return 0.
func (algo Algorithm) PackHashID() uint32 {
	return algo.info().packHashID
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

func (algo Algorithm) info() algorithmDetails {
	return algorithmTable[algo]
}
