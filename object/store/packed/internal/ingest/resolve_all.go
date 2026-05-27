package ingest

import (
	"errors"

	"lindenii.org/go/furgit/internal/progress"
)

// resolveAll resolves all delta records and finalizes ObjectID/RealType for every record.
func resolveAll(state *ingestState) error {
	state.unresolvedRefDeltas = state.unresolvedRefDeltas[:0]

	var pending uint32

	for idx := range state.records {
		if !state.records[idx].resolved {
			pending++
		}
	}

	if pending == 0 {
		return nil
	}

	var done uint32

	meter := progress.New(progress.Options{
		Writer: state.opts.Progress,
		Title:  "resolving deltas",
		Total:  uint64(pending),
	})

	for idx := range state.records {
		if state.records[idx].resolved {
			continue
		}

		done++
		meter.Set(uint64(done), 0)

		visiting := make(map[int]struct{})

		ty, content, err := resolveRecord(state, idx, visiting)
		if err != nil {
			if errors.Is(err, errExternalThinBase) {
				state.unresolvedRefDeltas = append(state.unresolvedRefDeltas, idx)

				continue
			}

			return err
		}

		id, err := hashCanonicalObject(state.algo, ty, content)
		if err != nil {
			return err
		}

		record := &state.records[idx]
		record.realType = ty
		record.objectID = id
		record.resolved = true
		state.objectToRecord[id] = idx
		state.baseCache.add(idx, ty, content)
	}

	meter.Stop("done")

	return nil
}
