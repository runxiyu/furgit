package ingest

import (
	"fmt"
	"io"

	"codeberg.org/lindenii/furgit/internal/compress/zlib"
	objectheader "codeberg.org/lindenii/furgit/object/header"
	objectid "codeberg.org/lindenii/furgit/object/id"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// drainEntryPayload inflates one entry payload from stream and returns
// (inflatedLength, oidForBaseEntry).
func drainEntryPayload(state *ingestState, record objectRecord) (int64, objectid.ObjectID, error) {
	var zero objectid.ObjectID

	reader, err := zlib.NewReader(state.stream)
	if err != nil {
		return 0, zero, &MalformedPackEntryError{Offset: record.offset, Reason: fmt.Sprintf("open zlib stream: %v", err)}
	}

	defer func() { _ = reader.Close() }()

	var total int64

	if record.packedType.IsBaseObject() {
		header, ok := objectheader.Encode(record.packedType, record.declaredSize)
		if !ok {
			return 0, zero, &MalformedPackEntryError{Offset: record.offset, Reason: "encode object header"}
		}

		hashImpl, err := state.algo.New()
		if err != nil {
			return 0, zero, err
		}

		_, _ = hashImpl.Write(header)

		n, err := io.Copy(hashImpl, reader)
		if err != nil {
			return 0, zero, &MalformedPackEntryError{Offset: record.offset, Reason: fmt.Sprintf("inflate base object: %v", err)}
		}

		total = n

		oid, err := objectid.FromBytes(state.algo, hashImpl.Sum(nil))
		if err != nil {
			return 0, zero, err
		}

		return total, oid, nil
	}

	if record.packedType == objecttype.TypeOfsDelta || record.packedType == objecttype.TypeRefDelta {
		n, err := io.Copy(io.Discard, reader)
		if err != nil {
			return 0, zero, &MalformedPackEntryError{Offset: record.offset, Reason: fmt.Sprintf("inflate delta payload: %v", err)}
		}

		total = n

		return total, zero, nil
	}

	return 0, zero, &MalformedPackEntryError{Offset: record.offset, Reason: "unsupported payload type"}
}
