package read

import (
	"encoding/binary"

	"codeberg.org/lindenii/furgit/format/commitgraph"
	"codeberg.org/lindenii/furgit/internal/intconv"
)

func (reader *Reader) readGenerationV2(layer *layer, index uint32, commitTime uint64) (uint64, error) {
	if len(layer.chunkGeneration) == 0 {
		return 0, nil
	}

	off64 := uint64(index) * 4

	off, err := intconv.Uint64ToInt(off64)
	if err != nil {
		return 0, err
	}

	value := binary.BigEndian.Uint32(layer.chunkGeneration[off : off+4])

	if value&commitgraph.GenerationOverflow == 0 {
		return commitTime + uint64(value), nil
	}

	gdo2Index := value ^ commitgraph.GenerationOverflow
	gdo2Off64 := uint64(gdo2Index) * 8

	gdo2Off, err := intconv.Uint64ToInt(gdo2Off64)
	if err != nil {
		return 0, err
	}

	if gdo2Off+8 > len(layer.chunkGenerationOv) {
		return 0, &MalformedError{Path: layer.path, Reason: "GDO2 index out of range"}
	}

	overflow := binary.BigEndian.Uint64(layer.chunkGenerationOv[gdo2Off : gdo2Off+8])

	return commitTime + overflow, nil
}
