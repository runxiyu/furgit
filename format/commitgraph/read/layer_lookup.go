package read

import (
	"bytes"
	"encoding/binary"

	"lindenii.org/go/furgit/internal/intconv"
	objectid "lindenii.org/go/furgit/object/id"
)

func layerLookup(layer *layer, oid objectid.ObjectID) (uint32, bool) {
	hashSize := oid.Algorithm().Size()
	first := int(oid.RawBytes()[0])

	var lo uint32
	if first > 0 {
		lo = binary.BigEndian.Uint32(layer.chunkOIDFanout[(first-1)*4 : first*4])
	}

	hi := binary.BigEndian.Uint32(layer.chunkOIDFanout[first*4 : (first+1)*4])
	if hi == 0 || lo >= hi {
		return 0, false
	}

	target := oid.RawBytes()
	left := int(lo)

	right := int(hi) - 1
	for left <= right {
		mid := left + (right-left)/2
		start := mid * hashSize
		end := start + hashSize

		current := layer.chunkOIDLookup[start:end]

		cmp := bytes.Compare(current, target)
		switch {
		case cmp == 0:
			pos, err := intconv.IntToUint32(mid)
			if err != nil {
				return 0, false
			}

			return pos, true
		case cmp < 0:
			left = mid + 1
		default:
			right = mid - 1
		}
	}

	return 0, false
}
