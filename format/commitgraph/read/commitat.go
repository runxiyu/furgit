package read

import (
	"encoding/binary"

	"codeberg.org/lindenii/furgit/internal/intconv"
	objectid "codeberg.org/lindenii/furgit/object/id"
)

// CommitAt returns decoded commit-graph metadata at one position.
//
// Labels: Life-Independent.
func (reader *Reader) CommitAt(pos Position) (Commit, error) {
	layer, err := reader.layerByPosition(pos)
	if err != nil {
		return Commit{}, err
	}

	hashSize := reader.algo.Size()
	stride := hashSize + 16

	strideU64, err := intconv.IntToUint64(stride)
	if err != nil {
		return Commit{}, err
	}

	start64 := uint64(pos.Index) * strideU64
	end64 := start64 + strideU64

	start, err := intconv.Uint64ToInt(start64)
	if err != nil {
		return Commit{}, err
	}

	end, err := intconv.Uint64ToInt(end64)
	if err != nil {
		return Commit{}, err
	}

	record := layer.chunkCommit[start:end]

	treeOID, err := objectid.FromBytes(reader.algo, record[:hashSize])
	if err != nil {
		return Commit{}, err
	}

	oid, err := reader.OIDAt(pos)
	if err != nil {
		return Commit{}, err
	}

	p1 := binary.BigEndian.Uint32(record[hashSize : hashSize+4])
	p2 := binary.BigEndian.Uint32(record[hashSize+4 : hashSize+8])
	genAndTimeHi := binary.BigEndian.Uint32(record[hashSize+8 : hashSize+12])
	timeLow := binary.BigEndian.Uint32(record[hashSize+12 : hashSize+16])

	timeHigh := uint64(genAndTimeHi & 0x3)
	commitTimeU64 := (timeHigh << 32) | uint64(timeLow)

	commitTime, err := intconv.Uint64ToInt64(commitTimeU64)
	if err != nil {
		return Commit{}, err
	}

	generationV1 := genAndTimeHi >> 2

	generationV2, err := reader.readGenerationV2(layer, pos.Index, commitTimeU64)
	if err != nil {
		return Commit{}, err
	}

	parent1, parent2, extra, err := reader.decodeParents(layer, p1, p2)
	if err != nil {
		return Commit{}, err
	}

	return Commit{
		OID:            oid,
		TreeOID:        treeOID,
		Parent1:        parent1,
		Parent2:        parent2,
		ExtraParents:   extra,
		CommitTimeUnix: commitTime,
		GenerationV1:   generationV1,
		GenerationV2:   generationV2,
	}, nil
}
