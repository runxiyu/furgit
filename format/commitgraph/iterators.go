package commitgraph

import (
	"iter"

	"codeberg.org/lindenii/furgit/internal/intconv"
	"codeberg.org/lindenii/furgit/objectid"
)

// AllPositions iterates all commit positions in native layer order.
func (reader *Reader) AllPositions() iter.Seq[Position] {
	return func(yield func(Position) bool) {
		for layerIdx := range reader.layers {
			layer := &reader.layers[layerIdx]

			graph, err := intconv.IntToUint32(layerIdx)
			if err != nil {
				return
			}

			for idx := range layer.numCommits {
				if !yield(Position{Graph: graph, Index: idx}) {
					return
				}
			}
		}
	}
}

// AllOIDs iterates all commit object IDs in native layer order.
func (reader *Reader) AllOIDs() iter.Seq[objectid.ObjectID] {
	return func(yield func(objectid.ObjectID) bool) {
		positions := reader.AllPositions()
		for pos := range positions {
			oid, err := reader.OIDAt(pos)
			if err != nil {
				return
			}

			if !yield(oid) {
				return
			}
		}
	}
}
