package read

import (
	"fmt"
	"os"

	objectid "lindenii.org/go/furgit/object/id"
)

// Open opens commit-graph data from one objects root.
//
// Labels: Deps-Borrowed.
func Open(root *os.Root, algo objectid.Algorithm, mode OpenMode) (*Reader, error) {
	if algo.Size() == 0 {
		return nil, objectid.ErrInvalidAlgorithm
	}

	switch mode {
	case OpenSingle:
		return openSingle(root, algo)
	case OpenChain:
		return openChain(root, algo)
	default:
		return nil, fmt.Errorf("commitgraph: invalid open mode %d", mode)
	}
}
