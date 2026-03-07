package ingest

import (
	"fmt"

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

	err = readAndValidatePackHeader(state)
	if err != nil {
		return err
	}

	state.records = make([]objectRecord, 0, state.objectCountHeader)
	state.ofsDeltas = make([]ofsDeltaRef, 0, state.objectCountHeader)
	state.refDeltas = make([]refDeltaRef, 0, state.objectCountHeader)

	for range state.objectCountHeader {
		nextOffset, err := scanOneEntry(state, state.stream.consumed)
		if err != nil {
			return err
		}

		if nextOffset != state.stream.consumed {
			return fmt.Errorf("format/pack/ingest: internal stream offset mismatch")
		}
	}

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
