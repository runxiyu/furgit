package id

import (
	"crypto/sha1" //#nosec:G505
	"crypto/sha256"
	"hash"
)

type algorithmDetails struct {
	name                string
	size                int
	sum                 func([]byte) ObjectID
	new                 func() hash.Hash
}

func (algo Algorithm) details() algorithmDetails {
	return algorithmTable[algo]
}

//nolint:gochecknoglobals
var algorithmTable = [...]algorithmDetails{
	AlgorithmUnknown: {}, //nolint:exhaustruct
	AlgorithmSHA1: {
		name:                "sha1",
		size:                sha1.Size,
		sum: func(data []byte) ObjectID {
			sum := sha1.Sum(data) //#nosec G401

			var id ObjectID
			copy(id.data[:], sum[:])
			id.algo = AlgorithmSHA1

			return id
		},
		new:       sha1.New,
	},
	AlgorithmSHA256: {
		name:                "sha256",
		size:                sha256.Size,
		sum: func(data []byte) ObjectID {
			sum := sha256.Sum256(data)

			var id ObjectID
			copy(id.data[:], sum[:])
			id.algo = AlgorithmSHA256

			return id
		},
		new:       sha256.New,
	},
}

// maxObjectIDSize MUST be >= the largest supported algorithm size.
const maxObjectIDSize = sha256.Size

var (
	//nolint:gochecknoglobals
	algorithmByName = map[string]Algorithm{}
	//nolint:gochecknoglobals
	supportedAlgorithms []Algorithm
)

func init() { //nolint:gochecknoinits
	// Skip over AlgorithmUnknown.
	for algo := Algorithm(1); int(algo) < len(algorithmTable); algo++ {
		info := &algorithmTable[algo]
		algorithmByName[info.name] = algo
		supportedAlgorithms = append(supportedAlgorithms, algo)
	}
}
