package commitgraph

import (
	"fmt"
	"os"

	"codeberg.org/lindenii/furgit/objectid"
)

// Open opens commit-graph data from one objects root.
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
		return nil, fmt.Errorf("format/commitgraph: invalid open mode %d", mode)
	}
}
