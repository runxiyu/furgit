package read

import (
	"encoding/binary"

	"codeberg.org/lindenii/furgit/commitgraph"
	"codeberg.org/lindenii/furgit/internal/intconv"
)

func (reader *Reader) decodeExtraEdgeList(layer *layer, edgeStart uint32) ([]Position, error) {
	if len(layer.chunkExtraEdges) == 0 {
		return nil, &MalformedError{Path: layer.path, Reason: "missing EDGE chunk"}
	}

	out := make([]Position, 0)

	cur := edgeStart
	for {
		off64 := uint64(cur) * 4

		off, err := intconv.Uint64ToInt(off64)
		if err != nil {
			return nil, err
		}

		if off+4 > len(layer.chunkExtraEdges) {
			return nil, &MalformedError{Path: layer.path, Reason: "EDGE index out of range"}
		}

		word := binary.BigEndian.Uint32(layer.chunkExtraEdges[off : off+4])
		parentGlobal := word & commitgraph.ParentLastMask

		parentPos, err := reader.globalToPosition(parentGlobal)
		if err != nil {
			return nil, err
		}

		out = append(out, parentPos)

		if word&commitgraph.ParentExtraMask != 0 {
			break
		}

		cur++
	}

	return out, nil
}
