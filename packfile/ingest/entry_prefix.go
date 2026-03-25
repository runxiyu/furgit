package ingest

import (
	"fmt"

	"codeberg.org/lindenii/furgit/internal/intconv"
	"codeberg.org/lindenii/furgit/objectid"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// parseEntryPrefix parses one entry prefix from stream.
func parseEntryPrefix(state *ingestState, startOffset uint64) (objectRecord, error) {
	var record objectRecord

	record.offset = startOffset

	first, err := state.stream.ReadByte()
	if err != nil {
		return record, &MalformedPackEntryError{Offset: startOffset, Reason: fmt.Sprintf("read first header byte: %v", err)}
	}

	record.packedType = objecttype.Type((first >> 4) & 0x07)
	size := int64(first & 0x0f)
	headerLen := uint32(1)
	shift := uint(4)
	b := first

	for b&0x80 != 0 {
		b, err = state.stream.ReadByte()
		if err != nil {
			return record, &MalformedPackEntryError{Offset: startOffset, Reason: fmt.Sprintf("read size continuation: %v", err)}
		}

		headerLen++
		size |= int64(b&0x7f) << shift
		shift += 7
	}

	if size < 0 {
		return record, &MalformedPackEntryError{Offset: startOffset, Reason: "negative declared size"}
	}

	record.declaredSize = size

	switch record.packedType {
	case objecttype.TypeCommit, objecttype.TypeTree, objecttype.TypeBlob, objecttype.TypeTag:
	case objecttype.TypeRefDelta:
		baseRaw := make([]byte, state.algo.Size())

		err := state.stream.readFull(baseRaw)
		if err != nil {
			return record, &MalformedPackEntryError{Offset: startOffset, Reason: fmt.Sprintf("read ref base: %v", err)}
		}

		baseID, err := objectid.FromBytes(state.algo, baseRaw)
		if err != nil {
			return record, &MalformedPackEntryError{Offset: startOffset, Reason: fmt.Sprintf("parse ref base: %v", err)}
		}

		record.baseObject = baseID

		baseRawLen, err := intconv.IntToUint32(len(baseRaw))
		if err != nil {
			return record, err
		}

		headerLen += baseRawLen
	case objecttype.TypeOfsDelta:
		dist, consumed, err := readOfsDistanceFromStream(state.stream)
		if err != nil {
			return record, &MalformedPackEntryError{Offset: startOffset, Reason: err.Error()}
		}

		if startOffset <= dist {
			return record, &MalformedPackEntryError{Offset: startOffset, Reason: "ofs base offset out of bounds"}
		}

		record.baseOffset = startOffset - dist

		consumedUint32, err := intconv.IntToUint32(consumed)
		if err != nil {
			return record, err
		}

		headerLen += consumedUint32
	case objecttype.TypeInvalid, objecttype.TypeFuture:
		return record, &MalformedPackEntryError{Offset: startOffset, Reason: fmt.Sprintf("unsupported object type %d", record.packedType)}
	default:
		return record, &MalformedPackEntryError{Offset: startOffset, Reason: fmt.Sprintf("unsupported object type %d", record.packedType)}
	}

	record.headerLen = headerLen

	return record, nil
}
