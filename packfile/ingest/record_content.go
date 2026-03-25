package ingest

import (
	"fmt"

	objecttype "codeberg.org/lindenii/furgit/object/type"
	packfmt "codeberg.org/lindenii/furgit/packfile"
)

// readBaseRecordContent reads canonical base content for one non-delta record.
func readBaseRecordContent(state *ingestState, idx int) (objecttype.Type, []byte, error) {
	record := state.records[idx]
	if !packfmt.IsBaseObjectType(record.packedType) {
		return objecttype.TypeInvalid, nil, fmt.Errorf("packfile/ingest: record %d is not a base object", idx)
	}

	content, err := inflateRecordPayload(state, idx)
	if err != nil {
		return objecttype.TypeInvalid, nil, err
	}

	if int64(len(content)) != record.declaredSize {
		return objecttype.TypeInvalid, nil, &MalformedPackEntryError{
			Offset: record.offset,
			Reason: fmt.Sprintf("base content size mismatch got %d want %d", len(content), record.declaredSize),
		}
	}

	return record.packedType, content, nil
}
