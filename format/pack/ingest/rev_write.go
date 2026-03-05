package ingest

import (
	"encoding/binary"
	"slices"

	"codeberg.org/lindenii/furgit/objectid"
)

const (
	revMagic   = 0x52494458
	revVersion = 1
)

// writeRev writes rev index for resolved records.
func writeRev(state *ingestState) error {
	if !state.writeRev {
		return nil
	}

	idxOrder := buildIdxOrder(state)

	recordToIdxPos := make([]int, len(state.records))
	for pos, recordIdx := range idxOrder {
		recordToIdxPos[recordIdx] = pos
	}

	packOrder := buildPackOrder(state)

	hashImpl, err := state.algo.New()
	if err != nil {
		return err
	}

	var scratch [8]byte
	binary.BigEndian.PutUint32(scratch[:4], revMagic)

	err = writeAndHash(state.revFile, hashImpl, scratch[:4])
	if err != nil {
		return err
	}

	binary.BigEndian.PutUint32(scratch[:4], revVersion)

	err = writeAndHash(state.revFile, hashImpl, scratch[:4])
	if err != nil {
		return err
	}

	binary.BigEndian.PutUint32(scratch[:4], hashID(state.algo))

	err = writeAndHash(state.revFile, hashImpl, scratch[:4])
	if err != nil {
		return err
	}

	for _, recordIdx := range packOrder {
		binary.BigEndian.PutUint32(scratch[:4], uint32(recordToIdxPos[recordIdx]))

		err := writeAndHash(state.revFile, hashImpl, scratch[:4])
		if err != nil {
			return err
		}
	}

	err = writeAndHash(state.revFile, hashImpl, state.packHash.Bytes())
	if err != nil {
		return err
	}

	revHash := hashImpl.Sum(nil)

	_, err = state.revFile.Write(revHash)
	if err != nil {
		return err
	}

	return state.revFile.Sync()
}

// buildPackOrder returns record indexes sorted by pack offset.
func buildPackOrder(state *ingestState) []int {
	out := make([]int, 0, len(state.records))
	for idx := range state.records {
		out = append(out, idx)
	}

	slices.SortFunc(out, func(a, b int) int {
		offA := state.records[a].offset

		offB := state.records[b].offset
		switch {
		case offA < offB:
			return -1
		case offA > offB:
			return 1
		default:
			return 0
		}
	})

	return out
}

// hashID converts object algorithm to pack hash-id encoding.
func hashID(algo objectid.Algorithm) uint32 {
	switch algo {
	case objectid.AlgorithmSHA1:
		return 1
	case objectid.AlgorithmSHA256:
		return 2
	default:
		return 0
	}
}
