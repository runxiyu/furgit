package packed

import (
	"os"

	"codeberg.org/lindenii/furgit/objectid"
)

// New creates a packed-object store rooted at an objects/pack directory.
func New(root *os.Root, algo objectid.Algorithm) (*Store, error) {
	if algo.Size() == 0 {
		return nil, objectid.ErrInvalidAlgorithm
	}

	return &Store{
		root:                root,
		algo:                algo,
		candidateByPack:     make(map[string]packCandidate),
		candidateNodeByPack: make(map[string]*packCandidateNode),
		idxByPack:           make(map[string]*idxFile),
		packs:               make(map[string]*packFile),
		deltaCache:          newDeltaCache(defaultDeltaCacheMaxBytes),
	}, nil
}
