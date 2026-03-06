package read

import "codeberg.org/lindenii/furgit/internal/intconv"

func (reader *Reader) layerByPosition(pos Position) (*layer, error) {
	graphIdx, err := intconv.Uint64ToInt(uint64(pos.Graph))
	if err != nil {
		return nil, err
	}

	if graphIdx < 0 || graphIdx >= len(reader.layers) {
		return nil, &PositionOutOfRangeError{Pos: pos}
	}

	layer := &reader.layers[graphIdx]
	if pos.Index >= layer.numCommits {
		return nil, &PositionOutOfRangeError{Pos: pos}
	}

	return layer, nil
}
