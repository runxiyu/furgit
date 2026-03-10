package read

import (
	"codeberg.org/lindenii/furgit/internal/intconv"
	"codeberg.org/lindenii/furgit/objectid"
)

// Lookup resolves one object ID to one graph position.
func (reader *Reader) Lookup(oid objectid.ObjectID) (Position, error) {
	if oid.Algorithm() != reader.algo {
		return Position{}, &NotFoundError{OID: oid}
	}

	for layerIdx := len(reader.layers) - 1; layerIdx >= 0; layerIdx-- {
		layer := &reader.layers[layerIdx]

		found, ok := layerLookup(layer, oid)
		if ok {
			idxU32, err := intconv.IntToUint32(layerIdx)
			if err != nil {
				return Position{}, err
			}

			return Position{Graph: idxU32, Index: found}, nil
		}
	}

	return Position{}, &NotFoundError{OID: oid}
}
