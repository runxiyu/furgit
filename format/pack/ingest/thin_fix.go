package ingest

import (
	"fmt"

	"codeberg.org/lindenii/furgit/internal/intconv"
	"codeberg.org/lindenii/furgit/internal/utils"
)

// maybeFixThin appends missing bases and rewrites pack header/trailer when needed.
func maybeFixThin(state *ingestState) error {
	if len(state.unresolvedRefDeltas) == 0 {
		return nil
	}

	utils.WriteProgressf(
		state.opts.Progress,
		"fixing thin pack: %d unresolved bases\r",
		len(state.unresolvedRefDeltas),
	)

	if !state.opts.FixThin {
		return &ThinPackUnresolvedError{Count: len(state.unresolvedRefDeltas)}
	}

	if state.opts.Base == nil {
		return &ThinPackUnresolvedError{Count: len(state.unresolvedRefDeltas)}
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
	total := len(baseIDs)
	if total > 0 {
		utils.WriteProgressf(state.opts.Progress, "fixing thin pack:   0%% (0/%d)\r", total)
	}

	for i, id := range baseIDs {
		ty, content, err := state.opts.Base.ReadBytesContent(id)
		if err != nil {
			continue
		}

		_, err = appendBaseObject(state, id, ty, content)
		if err != nil {
			return err
		}

		state.thinFixed = true

		done := i + 1
		percent := done * 100 / total
		utils.WriteProgressf(state.opts.Progress, "fixing thin pack: %3d%% (%d/%d)\r", percent, done, total)
	}

	err = rewritePackHeaderAndTrailer(state)
	if err != nil {
		return err
	}

	if state.thinFixed {
		utils.WriteProgressf(state.opts.Progress, "fixing thin pack: 100%% (%d/%d), done.\n", total, total)
	}

	return nil
}
