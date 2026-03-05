package ingest

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"slices"

	deltaapply "codeberg.org/lindenii/furgit/format/delta/apply"
	packfmt "codeberg.org/lindenii/furgit/format/pack"
	"codeberg.org/lindenii/furgit/internal/compress/zlib"
	"codeberg.org/lindenii/furgit/internal/intconv"
	"codeberg.org/lindenii/furgit/objectheader"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
)

var errExternalThinBase = errors.New("format/pack/ingest: external thin base required")

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

// resolveRecord resolves one record and returns canonical type/content.
func resolveRecord(state *ingestState, idx int, visiting map[int]struct{}) (objecttype.Type, []byte, error) {
	if idx < 0 || idx >= len(state.records) {
		return objecttype.TypeInvalid, nil, fmt.Errorf("format/pack/ingest: record index out of bounds")
	}

	if _, ok := visiting[idx]; ok {
		return objecttype.TypeInvalid, nil, &ErrDeltaCycle{Offset: state.records[idx].offset}
	}

	visiting[idx] = struct{}{}
	defer delete(visiting, idx)

	record := &state.records[idx]
	if ty, content, ok := state.baseCache.get(idx); ok {
		return ty, content, nil
	}

	if packfmt.IsBaseObjectType(record.packedType) {
		ty, content, err := readBaseRecordContent(state, idx)
		if err != nil {
			return objecttype.TypeInvalid, nil, err
		}

		if record.resolved {
			state.baseCache.add(idx, record.realType, content)

			return record.realType, content, nil
		}

		id, err := hashCanonicalObject(state.algo, ty, content)
		if err != nil {
			return objecttype.TypeInvalid, nil, err
		}

		record.objectID = id
		record.realType = ty
		record.resolved = true
		state.objectToRecord[id] = idx
		state.baseCache.add(idx, ty, content)

		return ty, content, nil
	}

	var (
		baseType    objecttype.Type
		baseContent []byte
		err         error
	)
	switch record.packedType {
	case objecttype.TypeOfsDelta:
		baseIdx, ok := state.offsetToRecord[record.baseOffset]
		if !ok {
			return objecttype.TypeInvalid, nil, &ErrMalformedPackEntry{
				Offset: record.offset,
				Reason: "missing ofs-delta base entry",
			}
		}

		baseType, baseContent, err = resolveRecord(state, baseIdx, visiting)
		if err != nil {
			return objecttype.TypeInvalid, nil, err
		}
	case objecttype.TypeRefDelta:
		baseIdx, ok := state.objectToRecord[record.baseObject]
		if ok {
			baseType, baseContent, err = resolveRecord(state, baseIdx, visiting)
			if err != nil {
				return objecttype.TypeInvalid, nil, err
			}
		} else {
			return objecttype.TypeInvalid, nil, errExternalThinBase
		}
	case objecttype.TypeInvalid,
		objecttype.TypeCommit,
		objecttype.TypeTree,
		objecttype.TypeBlob,
		objecttype.TypeTag,
		objecttype.TypeFuture:
		return objecttype.TypeInvalid, nil, &ErrMalformedPackEntry{
			Offset: record.offset,
			Reason: "unsupported delta type",
		}
	default:
		return objecttype.TypeInvalid, nil, &ErrMalformedPackEntry{
			Offset: record.offset,
			Reason: "unsupported delta type",
		}
	}

	ty, content, err := applyDeltaRecord(state, idx, baseType, baseContent)
	if err != nil {
		return objecttype.TypeInvalid, nil, err
	}

	id, err := hashCanonicalObject(state.algo, ty, content)
	if err != nil {
		return objecttype.TypeInvalid, nil, err
	}

	record.objectID = id
	record.realType = ty
	record.resolved = true
	state.objectToRecord[id] = idx
	state.baseCache.add(idx, ty, content)

	return ty, content, nil
}

// readBaseRecordContent reads canonical base content for one non-delta record.
func readBaseRecordContent(state *ingestState, idx int) (objecttype.Type, []byte, error) {
	record := state.records[idx]
	if !packfmt.IsBaseObjectType(record.packedType) {
		return objecttype.TypeInvalid, nil, fmt.Errorf("format/pack/ingest: record %d is not a base object", idx)
	}

	content, err := inflateRecordPayload(state, idx)
	if err != nil {
		return objecttype.TypeInvalid, nil, err
	}

	if int64(len(content)) != record.declaredSize {
		return objecttype.TypeInvalid, nil, &ErrMalformedPackEntry{
			Offset: record.offset,
			Reason: fmt.Sprintf("base content size mismatch got %d want %d", len(content), record.declaredSize),
		}
	}

	return record.packedType, content, nil
}

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
		return objecttype.TypeInvalid, nil, &ErrMalformedPackEntry{
			Offset: record.offset,
			Reason: fmt.Sprintf("delta payload size mismatch got %d want %d", len(deltaPayload), record.declaredSize),
		}
	}

	srcSize, dstSize, err := readDeltaHeaderSizes(deltaPayload)
	if err != nil {
		return objecttype.TypeInvalid, nil, &ErrMalformedPackEntry{
			Offset: record.offset,
			Reason: fmt.Sprintf("read delta header: %v", err),
		}
	}

	if srcSize != len(baseContent) {
		return objecttype.TypeInvalid, nil, &ErrMalformedPackEntry{
			Offset: record.offset,
			Reason: fmt.Sprintf("delta source size mismatch got %d want %d", srcSize, len(baseContent)),
		}
	}

	content, err := deltaapply.Apply(baseContent, deltaPayload)
	if err != nil {
		return objecttype.TypeInvalid, nil, &ErrMalformedPackEntry{
			Offset: record.offset,
			Reason: fmt.Sprintf("apply delta: %v", err),
		}
	}

	if len(content) != dstSize {
		return objecttype.TypeInvalid, nil, &ErrMalformedPackEntry{
			Offset: record.offset,
			Reason: fmt.Sprintf("delta result size mismatch got %d want %d", len(content), dstSize),
		}
	}

	return baseType, content, nil
}

// inflateRecordPayload inflates one record's zlib payload from pack file.
func inflateRecordPayload(state *ingestState, idx int) ([]byte, error) {
	record := state.records[idx]
	if record.packedLen < uint64(record.headerLen) {
		return nil, &ErrMalformedPackEntry{Offset: record.offset, Reason: "entry packed span underflow"}
	}

	compressedOffset := record.offset + uint64(record.headerLen)
	compressedLen := record.packedLen - uint64(record.headerLen)
	compressedOffsetInt64, err := intconv.Uint64ToInt64(compressedOffset)
	if err != nil {
		return nil, err
	}

	compressedLenInt64, err := intconv.Uint64ToInt64(compressedLen)
	if err != nil {
		return nil, err
	}

	section := io.NewSectionReader(state.packFile, compressedOffsetInt64, compressedLenInt64)

	reader, err := zlib.NewReader(section)
	if err != nil {
		return nil, &ErrMalformedPackEntry{Offset: record.offset, Reason: fmt.Sprintf("open payload zlib: %v", err)}
	}

	defer func() { _ = reader.Close() }()

	out, err := io.ReadAll(reader)
	if err != nil {
		return nil, &ErrMalformedPackEntry{Offset: record.offset, Reason: fmt.Sprintf("inflate payload: %v", err)}
	}

	return out, nil
}

// hashCanonicalObject hashes canonical object bytes (header+content).
func hashCanonicalObject(algo objectid.Algorithm, ty objecttype.Type, content []byte) (objectid.ObjectID, error) {
	header, ok := objectheader.Encode(ty, int64(len(content)))
	if !ok {
		return objectid.ObjectID{}, fmt.Errorf("format/pack/ingest: encode object header for type %d", ty)
	}

	hashImpl, err := algo.New()
	if err != nil {
		return objectid.ObjectID{}, err
	}

	_, _ = hashImpl.Write(header)
	_, _ = hashImpl.Write(content)

	return objectid.FromBytes(algo, hashImpl.Sum(nil))
}

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
