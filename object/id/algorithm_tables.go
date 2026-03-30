package objectid

import (
	"crypto/sha1" //#nosec:G505
	"crypto/sha256"
)

//nolint:gochecknoglobals
var algorithmTable = [...]algorithmDetails{
	AlgorithmUnknown: {},
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
		new: sha1.New,
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
		new: sha256.New,
	},
}

var (
	//nolint:gochecknoglobals
	algorithmByName = map[string]Algorithm{}
	//nolint:gochecknoglobals
	algorithmBySignatureHeaderName = map[string]Algorithm{}
	//nolint:gochecknoglobals
	supportedAlgorithms []Algorithm
)

func init() { //nolint:gochecknoinits
	emptyTreeInput := []byte("tree 0\x00")

	for algo := Algorithm(0); int(algo) < len(algorithmTable); algo++ {
		info := &algorithmTable[algo]
		if info.name == "" {
			continue
		}

		info.emptyTree = info.sum(emptyTreeInput)

		algorithmByName[info.name] = algo
		if info.signatureHeaderName != "" {
			algorithmBySignatureHeaderName[info.signatureHeaderName] = algo
		}

		supportedAlgorithms = append(supportedAlgorithms, algo)
	}
}
