package read

import (
	"os"

	"codeberg.org/lindenii/furgit/internal/intconv"
	"codeberg.org/lindenii/furgit/objectid"
)

func openSingle(root *os.Root, algo objectid.Algorithm) (*Reader, error) {
	graph, err := openLayer(root, "info/commit-graph", algo)
	if err != nil {
		return nil, err
	}

	graph.baseCount = 0
	graph.globalFrom = 0

	hashVersion, err := intconv.Uint32ToUint8(algo.PackHashID())
	if err != nil {
		return nil, err
	}

	out := &Reader{
		algo:        algo,
		hashVersion: hashVersion,
		layers:      []layer{*graph},
		total:       graph.numCommits,
	}

	return out, nil
}
