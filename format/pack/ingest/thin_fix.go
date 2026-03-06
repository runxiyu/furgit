package ingest

import (
	"fmt"

	"codeberg.org/lindenii/furgit/internal/intconv"
)

// maybeFixThin appends missing bases and rewrites pack header/trailer when needed.
func maybeFixThin(state *ingestState) error {
	if len(state.unresolvedRefDeltas) == 0 {
		return nil
	}

	if !state.fixThin {
		return &ErrThinPackUnresolved{Count: len(state.unresolvedRefDeltas)}
	}

	if state.base == nil {
		return &ErrThinPackUnresolved{Count: len(state.unresolvedRefDeltas)}
	}

	hashSize := int64(state.algo.Size())

	info, err := state.packFile.Stat()
	if err != nil {
		return err
	}

	size := info.Size()
	if size < hashSize {
		return fmt.Errorf("format/pack/ingest: pack too short to trim trailer")
	}

	newEnd := size - hashSize

	err = state.packFile.Truncate(newEnd)
	if err != nil {
		return err
	}

	consumed, err := intconv.Int64ToUint64(newEnd)
	if err != nil {
		return err
	}

	state.stream.consumed = consumed

	baseIDs := unresolvedThinBaseIDs(state)
	for _, id := range baseIDs {
		ty, content, err := state.base.ReadBytesContent(id)
		if err != nil {
			continue
		}

		_, err = appendBaseObject(state, id, ty, content)
		if err != nil {
			return err
		}

		state.thinFixed = true
	}

	err = rewritePackHeaderAndTrailer(state)
	if err != nil {
		return err
	}

	return nil
}
