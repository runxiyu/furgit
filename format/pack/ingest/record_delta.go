package ingest

import (
	"fmt"

	deltaapply "codeberg.org/lindenii/furgit/format/delta/apply"
	"codeberg.org/lindenii/furgit/objecttype"
)

// applyDeltaRecord applies one delta record onto base content.
func applyDeltaRecord(state *ingestState, idx int, baseType objecttype.Type, baseContent []byte) (objecttype.Type, []byte, error) {
	record := state.records[idx]
	if record.packedType != objecttype.TypeOfsDelta && record.packedType != objecttype.TypeRefDelta {
		return objecttype.TypeInvalid, nil, fmt.Errorf("format/pack/ingest: record %d is not a delta record", idx)
	}

	deltaPayload, err := inflateRecordPayload(state, idx)
	if err != nil {
		return objecttype.TypeInvalid, nil, err
	}

	if int64(len(deltaPayload)) != record.declaredSize {
		return objecttype.TypeInvalid, nil, &MalformedPackEntryError{
			Offset: record.offset,
			Reason: fmt.Sprintf("delta payload size mismatch got %d want %d", len(deltaPayload), record.declaredSize),
		}
	}

	srcSize, dstSize, err := readDeltaHeaderSizes(deltaPayload)
	if err != nil {
		return objecttype.TypeInvalid, nil, &MalformedPackEntryError{
			Offset: record.offset,
			Reason: fmt.Sprintf("read delta header: %v", err),
		}
	}

	if srcSize != len(baseContent) {
		return objecttype.TypeInvalid, nil, &MalformedPackEntryError{
			Offset: record.offset,
			Reason: fmt.Sprintf("delta source size mismatch got %d want %d", srcSize, len(baseContent)),
		}
	}

	content, err := deltaapply.Apply(baseContent, deltaPayload)
	if err != nil {
		return objecttype.TypeInvalid, nil, &MalformedPackEntryError{
			Offset: record.offset,
			Reason: fmt.Sprintf("apply delta: %v", err),
		}
	}

	if len(content) != dstSize {
		return objecttype.TypeInvalid, nil, &MalformedPackEntryError{
			Offset: record.offset,
			Reason: fmt.Sprintf("delta result size mismatch got %d want %d", len(content), dstSize),
		}
	}

	return baseType, content, nil
}
