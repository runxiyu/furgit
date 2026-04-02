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

func (algo Algorithm) info() algorithmDetails {
	return algorithmTable[algo]
}

//nolint:gochecknoglobals
var algorithmTable = [...]algorithmDetails{
	AlgorithmUnknown: {},
	AlgorithmSHA1: {
		name:                "sha1",
		size:                sha1.Size,
		packHashID:          1,
		signatureHeaderName: "gpgsig",
		new:                 sha1.New,
	},
	AlgorithmSHA256: {
		name:                "sha256",
		size:                sha256.Size,
		packHashID:          2,
		signatureHeaderName: "gpgsig-sha256",
		new:                 sha256.New,
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
