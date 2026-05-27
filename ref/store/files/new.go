package files

import (
	"math/rand"
	"os"
	"time"

	objectid "lindenii.org/go/furgit/object/id"
)

// New creates one files ref store rooted at one repository gitdir.
//
// Labels: Deps-Borrowed, Life-Parent.
func New(root *os.Root, algo objectid.Algorithm, packedRefsTimeout time.Duration) (*Store, error) {
	if algo.Size() == 0 {
		return nil, objectid.ErrInvalidAlgorithm
	}

	commonRoot, err := openCommonRoot(root)
	if err != nil {
		return nil, err
	}

	return &Store{
		gitRoot:           root,
		commonRoot:        commonRoot,
		algo:              algo,
		lockRand:          rand.New(rand.NewSource(time.Now().UnixNano())), //nolint:gosec
		packedRefsTimeout: packedRefsTimeout,
	}, nil
}
