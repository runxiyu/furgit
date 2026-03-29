package read

import (
	"iter"

	"codeberg.org/lindenii/furgit/internal/intconv"
	objectid "codeberg.org/lindenii/furgit/object/id"
)

// AllPositions iterates all commit positions in native layer order.
//
// Labels: MT-Safe, Life-Parent.
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
//
// Labels: MT-Safe, Life-Parent.
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
