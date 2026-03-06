package read

import (
	"codeberg.org/lindenii/furgit/internal/intconv"
	"codeberg.org/lindenii/furgit/objectid"
)

// OIDAt returns object ID at one position.
func (reader *Reader) OIDAt(pos Position) (objectid.ObjectID, error) {
	layer, err := reader.layerByPosition(pos)
	if err != nil {
		return objectid.ObjectID{}, err
	}

	hashSize := reader.algo.Size()

	hashSizeU64, err := intconv.IntToUint64(hashSize)
	if err != nil {
		return objectid.ObjectID{}, err
	}

	start64 := uint64(pos.Index) * hashSizeU64
	end64 := start64 + hashSizeU64

	start, err := intconv.Uint64ToInt(start64)
	if err != nil {
		return objectid.ObjectID{}, err
	}

	end, err := intconv.Uint64ToInt(end64)
	if err != nil {
		return objectid.ObjectID{}, err
	}

	return objectid.FromBytes(reader.algo, layer.chunkOIDLookup[start:end])
}
