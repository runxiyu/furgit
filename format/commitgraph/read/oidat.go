package read

import (
	"lindenii.org/go/furgit/internal/intconv"
	objectid "lindenii.org/go/furgit/object/id"
)

// OIDAt returns object ID at one position.
//
// Labels: Life-Independent.
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
