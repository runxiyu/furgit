package ingest

import (
	"errors"

	"codeberg.org/lindenii/furgit/internal/utils"
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

	step := progressStep(pending)
	var done uint32
	utils.WriteProgressf(state.opts.Progress, "resolving deltas:   0%% (0/%d)\r", pending)

	for idx := range state.records {
		if state.records[idx].resolved {
			continue
		}

		done++
		if done%step == 0 || done == pending {
			percent := done * 100 / pending
			utils.WriteProgressf(state.opts.Progress, "resolving deltas: %3d%% (%d/%d)\r", percent, done, pending)
		}

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

	utils.WriteProgressf(state.opts.Progress, "resolving deltas: 100%% (%d/%d), done.\n", pending, pending)

	return nil
}
