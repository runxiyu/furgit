package ingest

import (
	"errors"
	"fmt"

	"codeberg.org/lindenii/furgit/internal/intconv"
	"codeberg.org/lindenii/furgit/internal/progress"
	objectstore "codeberg.org/lindenii/furgit/object/store"
)

// maybeFixThin appends missing bases and rewrites pack header/trailer when needed.
func maybeFixThin(state *ingestState) error {
	if len(state.unresolvedRefDeltas) == 0 {
		return nil
	}

	writeProgressf(
		state,
		"fixing thin pack: %d unresolved bases\r",
		len(state.unresolvedRefDeltas),
	)

	if !state.opts.FixThin {
		return &ThinPackUnresolvedError{Count: len(state.unresolvedRefDeltas)}
	}

	if state.opts.ThinBase == nil {
		return &ThinPackUnresolvedError{Count: len(state.unresolvedRefDeltas)}
	}

	hashSize := int64(state.algo.Size())

	info, err := state.packFile.Stat()
	if err != nil {
		return err
	}

	size := info.Size()
	if size < hashSize {
		return fmt.Errorf("packfile/ingest: pack too short to trim trailer")
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
	meter := progress.New(progress.Options{
		Writer: state.opts.Progress,
		Title:  "fixing thin pack",
		Total:  uint64(total),
	})
	meter.Set(0, 0)

	var appended uint64

	for _, id := range baseIDs {
		ty, content, err := state.opts.ThinBase.ReadBytesContent(id)
		if err != nil {
			if errors.Is(err, objectstore.ErrObjectNotFound) {
				continue
			}

			return fmt.Errorf("packfile/ingest: read thin base %s: %w", id, err)
		}

		_, err = appendBaseObject(state, id, ty, content)
		if err != nil {
			return err
		}

		state.thinFixed = true

		appended++
		meter.Set(appended, 0)
	}

	err = rewritePackHeaderAndTrailer(state)
	if err != nil {
		return err
	}

	meter.Stop(fmt.Sprintf("appended %d/%d, done", appended, total))

	return nil
}
