package ingest

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash"
	"io"
	"slices"

	"codeberg.org/lindenii/furgit/internal/intconv"
)

const (
	idxMagicV2   = 0xff744f63
	idxVersionV2 = 2
)

// writeIdx writes idx v2 for resolved records.
func writeIdx(state *ingestState) error {
	order := buildIdxOrder(state)

	hashImpl, err := state.algo.New()
	if err != nil {
		return err
	}

	write := func(src []byte) error {
		_, err := state.idxFile.Write(src)
		if err != nil {
			return err
		}

		_, err = hashImpl.Write(src)
		if err != nil {
			return err
		}

		return nil
	}

	var scratch [8]byte
	binary.BigEndian.PutUint32(scratch[:4], idxMagicV2)
	binary.BigEndian.PutUint32(scratch[4:8], idxVersionV2)

	err = write(scratch[:8])
	if err != nil {
		return err
	}

	var fanout [256]uint32

	for _, recordIdx := range order {
		idRaw := state.records[recordIdx].objectID.Bytes()
		fanout[idRaw[0]]++
	}

	var cumulative uint32
	for i := range fanout {
		cumulative += fanout[i]
		binary.BigEndian.PutUint32(scratch[:4], cumulative)

		err := write(scratch[:4])
		if err != nil {
			return err
		}
	}

	for _, recordIdx := range order {
		idRaw := state.records[recordIdx].objectID.Bytes()

		err := write(idRaw)
		if err != nil {
			return err
		}
	}

	for _, recordIdx := range order {
		binary.BigEndian.PutUint32(scratch[:4], state.records[recordIdx].crc32)

		err := write(scratch[:4])
		if err != nil {
			return err
		}
	}

	largeOffsets := make([]uint64, 0)

	for _, recordIdx := range order {
		offset := state.records[recordIdx].offset
		if offset >= 0x80000000 {
			largeOffsetIdx, err := intconv.IntToUint32(len(largeOffsets))
			if err != nil {
				return err
			}

			word := 0x80000000 | largeOffsetIdx

			largeOffsets = append(largeOffsets, offset)

			binary.BigEndian.PutUint32(scratch[:4], word)
		} else {
			binary.BigEndian.PutUint32(scratch[:4], uint32(offset))
		}

		err := write(scratch[:4])
		if err != nil {
			return err
		}
	}

	for _, off := range largeOffsets {
		binary.BigEndian.PutUint64(scratch[:8], off)

		err := write(scratch[:8])
		if err != nil {
			return err
		}
	}

	err = write(state.packHash.Bytes())
	if err != nil {
		return err
	}

	idxHash := hashImpl.Sum(nil)

	_, err = state.idxFile.Write(idxHash)
	if err != nil {
		return err
	}

	return state.idxFile.Sync()
}

// buildIdxOrder returns record indexes sorted by ObjectID.
func buildIdxOrder(state *ingestState) []int {
	out := make([]int, 0, len(state.records))
	for idx := range state.records {
		out = append(out, idx)
	}

	slices.SortFunc(out, func(a, b int) int {
		return bytes.Compare(state.records[a].objectID.Bytes(), state.records[b].objectID.Bytes())
	})

	return out
}

// verifyResolvedRecords checks that all records are fully resolved before index writing.
func verifyResolvedRecords(state *ingestState) error {
	for idx, record := range state.records {
		if !record.resolved {
			return fmt.Errorf("format/pack/ingest: unresolved record %d at offset %d", idx, record.offset)
		}
	}

	return nil
}

// writeAndHash writes src to dst and updates hash.
func writeAndHash(dst io.Writer, hashImpl hash.Hash, src []byte) error {
	_, err := dst.Write(src)
	if err != nil {
		return err
	}

	_, err = hashImpl.Write(src)
	if err != nil {
		return err
	}

	return nil
}
