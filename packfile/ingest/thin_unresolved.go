package ingest

import (
	"bytes"
	"slices"

	"codeberg.org/lindenii/furgit/objectid"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// unresolvedThinBaseIDs returns sorted unique unresolved ref base IDs.
func unresolvedThinBaseIDs(state *ingestState) []objectid.ObjectID {
	seen := make(map[objectid.ObjectID]struct{})

	for _, idx := range state.unresolvedRefDeltas {
		record := state.records[idx]
		if record.packedType != objecttype.TypeRefDelta {
			continue
		}

		seen[record.baseObject] = struct{}{}
	}

	out := make([]objectid.ObjectID, 0, len(seen))
	for id := range seen {
		out = append(out, id)
	}

	slices.SortFunc(out, func(a, b objectid.ObjectID) int {
		return bytes.Compare(a.RawBytes(), b.RawBytes())
	})

	return out
}
