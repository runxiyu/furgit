package ingest

import (
	"fmt"

	"codeberg.org/lindenii/furgit/internal/utils"
	"codeberg.org/lindenii/furgit/objectid"
)

// streamPackAndScan copies src into temp .pack while scanning packed entries.
func streamPackAndScan(state *ingestState) error {
	hashImpl, err := state.algo.New()
	if err != nil {
		return err
	}

	state.stream = newStreamScanner(
		state.src,
		state.packFile,
		hashImpl,
		state.algo.Size(),
	)

	utils.WriteProgressf(state.opts.Progress, "validating pack header...\r")
	err = readAndValidatePackHeader(state)
	if err != nil {
		return err
	}
	utils.WriteProgressf(state.opts.Progress, "validating pack header: done.\n")

	state.records = make([]objectRecord, 0, state.objectCountHeader)
	state.ofsDeltas = make([]ofsDeltaRef, 0, state.objectCountHeader)
	state.refDeltas = make([]refDeltaRef, 0, state.objectCountHeader)

	total := state.objectCountHeader
	step := progressStep(total)
	utils.WriteProgressf(state.opts.Progress, "receiving objects:   0%% (0/%d)\r", total)

	for i := uint32(0); i < total; i++ {
		nextOffset, err := scanOneEntry(state, state.stream.consumed)
		if err != nil {
			return err
		}

		if nextOffset != state.stream.consumed {
			return fmt.Errorf("format/pack/ingest: internal stream offset mismatch")
		}

		done := i + 1
		if done%step == 0 || done == total {
			percent := done * 100 / total
			utils.WriteProgressf(state.opts.Progress, "receiving objects: %3d%% (%d/%d)\r", percent, done, total)
		}
	}

	utils.WriteProgressf(state.opts.Progress, "receiving objects: 100%% (%d/%d), done.\n", total, total)

	err = state.stream.finishAndFlushTrailer(state.opts.RequireTrailingEOF)
	if err != nil {
		return err
	}

	if len(state.stream.packTrailer) != state.algo.Size() {
		return fmt.Errorf("format/pack/ingest: invalid trailer size")
	}

	packHash, err := objectid.FromBytes(state.algo, state.stream.packTrailer)
	if err != nil {
		return err
	}

	state.packHash = packHash

	return state.stream.flush()
}
