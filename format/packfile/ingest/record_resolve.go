package ingest

import (
	"fmt"

	objecttype "codeberg.org/lindenii/furgit/object/type"
	packfmt "codeberg.org/lindenii/furgit/format/packfile"
)

// resolveRecord resolves one record and returns canonical type/content.
func resolveRecord(state *ingestState, idx int, visiting map[int]struct{}) (objecttype.Type, []byte, error) {
	if idx < 0 || idx >= len(state.records) {
		return objecttype.TypeInvalid, nil, fmt.Errorf("packfile/ingest: record index out of bounds")
	}

	if _, ok := visiting[idx]; ok {
		return objecttype.TypeInvalid, nil, &DeltaCycleError{Offset: state.records[idx].offset}
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
			return objecttype.TypeInvalid, nil, &MalformedPackEntryError{
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
		return objecttype.TypeInvalid, nil, &MalformedPackEntryError{
			Offset: record.offset,
			Reason: "unsupported delta type",
		}
	default:
		return objecttype.TypeInvalid, nil, &MalformedPackEntryError{
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
