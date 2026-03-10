package ingest

import (
	"compress/zlib"
	"hash/crc32"
	"io"

	"codeberg.org/lindenii/furgit/internal/intconv"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
)

// appendBaseObject appends one base object as a new packed non-delta entry.
func appendBaseObject(state *ingestState, id objectid.ObjectID, realType objecttype.Type, content []byte) (int, error) {
	start := state.stream.consumed

	header := encodePackEntryHeader(realType, int64(len(content)))

	startInt64, err := intconv.Uint64ToInt64(start)
	if err != nil {
		return 0, err
	}

	_, err = state.packFile.WriteAt(header, startInt64)
	if err != nil {
		return 0, err
	}

	headerLenInt64 := int64(len(header))
	section := &fileSectionWriter{file: state.packFile, off: startInt64 + headerLenInt64}
	crc := crc32.NewIEEE()

	_, err = crc.Write(header)
	if err != nil {
		return 0, err
	}

	counting := &countingWriter{dst: section}

	zw := zlib.NewWriter(io.MultiWriter(counting, crc))

	_, err = zw.Write(content)
	if err != nil {
		return 0, err
	}

	err = zw.Close()
	if err != nil {
		return 0, err
	}

	headerLenUint64, err := intconv.IntToUint64(len(header))
	if err != nil {
		return 0, err
	}

	countingNUint64, err := intconv.IntToUint64(counting.n)
	if err != nil {
		return 0, err
	}

	packedLen := headerLenUint64 + countingNUint64
	end := start + packedLen
	state.stream.consumed = end

	headerLenUint32, err := intconv.IntToUint32(len(header))
	if err != nil {
		return 0, err
	}

	record := objectRecord{
		offset:       start,
		headerLen:    headerLenUint32,
		packedLen:    packedLen,
		crc32:        crc.Sum32(),
		packedType:   realType,
		realType:     realType,
		declaredSize: int64(len(content)),
		dataOffset:   start + headerLenUint64,
		objectID:     id,
		resolved:     true,
	}

	recordIdx := len(state.records)
	state.records = append(state.records, record)
	state.offsetToRecord[start] = recordIdx
	state.objectToRecord[id] = recordIdx
	state.baseCache.add(recordIdx, realType, content)

	return recordIdx, nil
}
