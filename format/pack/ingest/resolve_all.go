package ingest

import "errors"

// resolveAll resolves all delta records and finalizes ObjectID/RealType for every record.
func resolveAll(state *ingestState) error {
	state.unresolvedRefDeltas = state.unresolvedRefDeltas[:0]

	for idx := range state.records {
		if state.records[idx].resolved {
			continue
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

	return nil
}
