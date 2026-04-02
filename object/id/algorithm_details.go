package id

import (
	"crypto/sha1" //#nosec:G505
	"crypto/sha256"
	"hash"
)

type algorithmDetails struct {
	name                string
	size                int
	packHashID          uint32
	signatureHeaderName string
	sum                 func([]byte) ObjectID
	new                 func() hash.Hash
	emptyTree           ObjectID
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
		packHashID:          1,
		signatureHeaderName: "gpgsig",
		sum: func(data []byte) ObjectID {
			sum := sha1.Sum(data) //#nosec G401

			var id ObjectID
			copy(id.data[:], sum[:])
			id.algo = AlgorithmSHA1

			return id
		},
		new:       sha1.New,
		emptyTree: ObjectID{}, //nolint:exhaustruct
	},
	AlgorithmSHA256: {
		name:                "sha256",
		size:                sha256.Size,
		packHashID:          2,
		signatureHeaderName: "gpgsig-sha256",
		sum: func(data []byte) ObjectID {
			sum := sha256.Sum256(data)

			var id ObjectID
			copy(id.data[:], sum[:])
			id.algo = AlgorithmSHA256

			return id
		},
		new:       sha256.New,
		emptyTree: ObjectID{}, //nolint:exhaustruct
	},
}

// maxObjectIDSize MUST be >= the largest supported algorithm size.
const maxObjectIDSize = sha256.Size

var (
	//nolint:gochecknoglobals
	algorithmByName = map[string]Algorithm{}
	//nolint:gochecknoglobals
	algorithmBySignatureHeaderName = map[string]Algorithm{}
	//nolint:gochecknoglobals
	supportedAlgorithms []Algorithm
)

func init() { //nolint:gochecknoinits
	// Skip over AlgorithmUnknown.
	for algo := Algorithm(1); int(algo) < len(algorithmTable); algo++ {
		info := &algorithmTable[algo]

		info.emptyTree = info.sum([]byte("tree 0\x00"))

		algorithmByName[info.name] = algo
		if info.signatureHeaderName != "" {
			algorithmBySignatureHeaderName[info.signatureHeaderName] = algo
		}

		supportedAlgorithms = append(supportedAlgorithms, algo)
	}
}
