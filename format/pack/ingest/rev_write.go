package ingest

import (
	"encoding/binary"
	"slices"

	"codeberg.org/lindenii/furgit/internal/intconv"
	"codeberg.org/lindenii/furgit/internal/progress"
)

const (
	revMagic   = 0x52494458
	revVersion = 1
)

// writeRev writes rev index for resolved records.
func writeRev(state *ingestState) error {
	if !state.opts.WriteRev {
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
	writeProgress(state, "writing reverse index header...\r")
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

	binary.BigEndian.PutUint32(scratch[:4], state.algo.PackHashID())

	err = writeAndHash(state.revFile, hashImpl, scratch[:4])
	if err != nil {
		return err
	}
	writeProgress(state, "writing reverse index header: done.\n")

	entriesMeter := progress.New(progress.Options{
		Writer: state.opts.Progress,
		Flush:  state.opts.ProgressFlush,
		Title:  "writing reverse index entries",
		Total:  uint64(len(packOrder)),
	})
	var entriesDone uint64

	for _, recordIdx := range packOrder {
		recordPos, err := intconv.IntToUint32(recordToIdxPos[recordIdx])
		if err != nil {
			return err
		}

		binary.BigEndian.PutUint32(scratch[:4], recordPos)

		err = writeAndHash(state.revFile, hashImpl, scratch[:4])
		if err != nil {
			return err
		}
		entriesDone++
		entriesMeter.Set(entriesDone, 0)
	}
	if entriesDone > 0 {
		entriesMeter.Stop("done")
	}

	writeProgress(state, "writing reverse index trailer...\r")

	err = writeAndHash(state.revFile, hashImpl, state.packHash.Bytes())
	if err != nil {
		return err
	}

	revHash := hashImpl.Sum(nil)

	_, err = state.revFile.Write(revHash)
	if err != nil {
		return err
	}

	err = state.revFile.Sync()
	if err != nil {
		return err
	}
	writeProgress(state, "writing reverse index trailer: done.\n")

	return nil
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
