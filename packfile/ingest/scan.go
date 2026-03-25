package ingest

import (
	"fmt"

	"codeberg.org/lindenii/furgit/internal/progress"
	objectid "codeberg.org/lindenii/furgit/object/id"
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

	writeProgressf(state, "validating pack header...\r")

	err = seedStreamWithPackHeader(state)
	if err != nil {
		return err
	}

	writeProgressf(state, "validating pack header: done.\n")

	state.records = make([]objectRecord, 0, state.objectCountHeader)
	state.ofsDeltas = make([]ofsDeltaRef, 0, state.objectCountHeader)
	state.refDeltas = make([]refDeltaRef, 0, state.objectCountHeader)

	total := state.objectCountHeader
	meter := progress.New(progress.Options{
		Writer:     state.opts.Progress,
		Flush:      state.opts.ProgressFlush,
		Title:      "receiving objects",
		Total:      uint64(total),
		Throughput: true,
	})

	for i := range total {
		nextOffset, err := scanOneEntry(state, state.stream.consumed)
		if err != nil {
			return err
		}

		if nextOffset != state.stream.consumed {
			return fmt.Errorf("packfile/ingest: internal stream offset mismatch")
		}

		done := i + 1
		meter.Set(uint64(done), state.stream.consumed)
	}

	meter.Stop("done")

	err = state.stream.finishAndFlushTrailer(state.opts.RequireTrailingEOF)
	if err != nil {
		return err
	}

	if len(state.stream.packTrailer) != state.algo.Size() {
		return fmt.Errorf("packfile/ingest: invalid trailer size")
	}

	packHash, err := objectid.FromBytes(state.algo, state.stream.packTrailer)
	if err != nil {
		return err
	}

	state.packHash = packHash

	return state.stream.flush()
}

// seedStreamWithPackHeader writes the already-validated PACK header to output,
// seeds the running pack hash, and advances stream offset accounting.
func seedStreamWithPackHeader(state *ingestState) error {
	written := 0
	for written < len(state.packHeaderRaw) {
		n, err := state.packFile.Write(state.packHeaderRaw[written:])
		if err != nil {
			return &DestinationWriteError{Op: fmt.Sprintf("write pack header: %v", err)}
		}

		if n == 0 {
			return &DestinationWriteError{Op: "write pack header: short write"}
		}

		written += n
	}

	_, err := state.stream.hash.Write(state.packHeaderRaw[:])
	if err != nil {
		return err
	}

	state.stream.consumed = packHeaderSize

	return nil
}
