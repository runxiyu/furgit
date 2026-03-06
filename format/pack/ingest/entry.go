package ingest

import (
	"fmt"

	packfmt "codeberg.org/lindenii/furgit/format/pack"
	"codeberg.org/lindenii/furgit/objecttype"
)

// scanOneEntry scans one pack entry from stream and appends one record.
func scanOneEntry(state *ingestState, startOffset uint64) (uint64, error) {
	state.stream.beginEntryCRC()

	record, err := parseEntryPrefix(state, startOffset)
	if err != nil {
		return 0, err
	}

	contentLen, consumedInput, oid, err := drainEntryPayload(state, record)
	if err != nil {
		return 0, err
	}

	if contentLen != record.declaredSize {
		return 0, &MalformedPackEntryError{
			Offset: startOffset,
			Reason: fmt.Sprintf("inflated size mismatch got %d want %d", contentLen, record.declaredSize),
		}
	}

	endOffset := startOffset + uint64(record.headerLen) + consumedInput
	if endOffset > state.stream.consumed {
		return 0, &MalformedPackEntryError{
			Offset: startOffset,
			Reason: fmt.Sprintf("entry end offset overflow got %d > stream %d", endOffset, state.stream.consumed),
		}
	}

	record.packedLen = endOffset - startOffset

	record.dataOffset = startOffset + uint64(record.headerLen)
	if record.packedLen < uint64(record.headerLen) {
		return 0, &MalformedPackEntryError{Offset: startOffset, Reason: "negative payload span"}
	}

	crc, err := state.stream.endEntryCRC()
	if err != nil {
		return 0, err
	}

	record.crc32 = crc

	if packfmt.IsBaseObjectType(record.packedType) {
		record.objectID = oid
		record.realType = record.packedType
		record.resolved = true
	}

	recordIdx := len(state.records)
	state.records = append(state.records, record)

	state.offsetToRecord[record.offset] = recordIdx
	if record.resolved {
		state.objectToRecord[record.objectID] = recordIdx
	}

	switch record.packedType {
	case objecttype.TypeOfsDelta:
		state.ofsDeltas = append(state.ofsDeltas, ofsDeltaRef{
			baseOffset: record.baseOffset,
			recordIdx:  recordIdx,
		})
	case objecttype.TypeRefDelta:
		state.refDeltas = append(state.refDeltas, refDeltaRef{
			baseObject: record.baseObject,
			recordIdx:  recordIdx,
		})
	case objecttype.TypeInvalid,
		objecttype.TypeCommit,
		objecttype.TypeTree,
		objecttype.TypeBlob,
		objecttype.TypeTag,
		objecttype.TypeFuture:
	default:
	}

	return endOffset, nil
}
