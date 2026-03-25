package packed

import (
	"fmt"
	"os"

	objectid "codeberg.org/lindenii/furgit/object/id"
)

// New creates a packed-object store rooted at an objects/pack directory.
func New(root *os.Root, algo objectid.Algorithm, opts Options) (*Store, error) {
	if algo.Size() == 0 {
		return nil, objectid.ErrInvalidAlgorithm
	}

	switch opts.RefreshPolicy {
	case RefreshPolicyOnMissing, RefreshPolicyNever:
	default:
		return nil, fmt.Errorf("objectstorer/packed: invalid refresh policy %d", opts.RefreshPolicy)
	}

	return &Store{
		root:          root,
		algo:          algo,
		refreshPolicy: opts.RefreshPolicy,
		mruNodeByPack: make(map[string]*packCandidateNode),
		idxByPack:     make(map[string]*idxFile),
		packs:         make(map[string]*packFile),
		deltaCache:    newDeltaCache(defaultDeltaCacheMaxBytes),
	}, nil
}
