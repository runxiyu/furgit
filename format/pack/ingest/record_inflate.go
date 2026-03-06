package ingest

import (
	"compress/zlib"
	"fmt"
	"io"

	"codeberg.org/lindenii/furgit/internal/intconv"
)

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
