package ingest

import (
	"errors"
	"fmt"

	"codeberg.org/lindenii/furgit/internal/intconv"
	"codeberg.org/lindenii/furgit/internal/progress"
	"codeberg.org/lindenii/furgit/objectstore"
)

// maybeFixThin appends missing bases and rewrites pack header/trailer when needed.
func maybeFixThin(state *ingestState) error {
	if len(state.unresolvedRefDeltas) == 0 {
		return nil
	}

	writeProgress(
		state,
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
	meter := progress.New(progress.Options{
		Writer: state.opts.Progress,
		Flush:  state.opts.ProgressFlush,
		Title:  "fixing thin pack",
		Total:  uint64(total),
	})

	var appended uint64

	for _, id := range baseIDs {
		ty, content, err := state.opts.Base.ReadBytesContent(id)
		if err != nil {
			if errors.Is(err, objectstore.ErrObjectNotFound) {
				continue
			}

			return fmt.Errorf("format/pack/ingest: read thin base %s: %w", id, err)
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

	if state.thinFixed {
		meter.Stop("done")
	}

	return nil
}
