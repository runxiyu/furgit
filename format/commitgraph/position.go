package commitgraph

import (
	"fmt"

	"codeberg.org/lindenii/furgit/internal/intconv"
)

// Position identifies one commit record by layer and row index.
type Position struct {
	Graph uint32
	Index uint32
}

func (reader *Reader) globalToPosition(global uint32) (Position, error) {
	for i := range reader.layers {
		layer := &reader.layers[i]
		from := layer.globalFrom

		to := from + layer.numCommits
		if global >= from && global < to {
			graph, err := intconv.IntToUint32(i)
			if err != nil {
				return Position{}, err
			}

			return Position{
				Graph: graph,
				Index: global - from,
			}, nil
		}
	}

	return Position{}, &ErrMalformed{
		Path:   "commit-graph",
		Reason: fmt.Sprintf("parent global position out of range: %d", global),
	}
}
