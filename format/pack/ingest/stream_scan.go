package ingest

import (
	"encoding/binary"
	"fmt"
	"io"

	deltaapply "codeberg.org/lindenii/furgit/format/delta/apply"
	packfmt "codeberg.org/lindenii/furgit/format/pack"
	"codeberg.org/lindenii/furgit/internal/compress/zlib"
	"codeberg.org/lindenii/furgit/objectheader"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
)

// streamPackAndScan copies src into temp .pack while scanning packed entries.
func streamPackAndScan(state *ingestState) error {
	hashImpl, err := state.algo.New()
	if err != nil {
		return err
	}

	state.stream = newStreamScanner(
		state.src,
		state.packFile,
		hashImpl,
		state.algo.Size(),
	)

	if err := readAndValidatePackHeader(state); err != nil {
		return err
	}

	state.records = make([]objectRecord, 0, state.objectCountHeader)
	state.ofsDeltas = make([]ofsDeltaRef, 0, state.objectCountHeader)
	state.refDeltas = make([]refDeltaRef, 0, state.objectCountHeader)

	for range state.objectCountHeader {
		nextOffset, err := scanOneEntry(state, state.stream.consumed)
		if err != nil {
			return err
		}

		if nextOffset != state.stream.consumed {
			return fmt.Errorf("format/pack/ingest: internal stream offset mismatch")
		}
	}

	if err := state.stream.finishAndFlushTrailer(); err != nil {
		return err
	}
	if len(state.stream.packTrailer) != state.algo.Size() {
		return fmt.Errorf("format/pack/ingest: invalid trailer size")
	}
	packHash, err := objectid.FromBytes(state.algo, state.stream.packTrailer)
	if err != nil {
		return err
	}
	state.packHash = packHash

	return state.stream.flush()
}

// readAndValidatePackHeader reads and validates PACK header from the stream.
func readAndValidatePackHeader(state *ingestState) error {
	var hdr [12]byte
	if err := state.stream.readFull(hdr[:]); err != nil {
		return &ErrInvalidPackHeader{Reason: fmt.Sprintf("read header: %v", err)}
	}

	if binary.BigEndian.Uint32(hdr[:4]) != packfmt.Signature {
		return &ErrInvalidPackHeader{Reason: "signature mismatch"}
	}

	version := binary.BigEndian.Uint32(hdr[4:8])
	if !packfmt.VersionSupported(version) {
		return &ErrInvalidPackHeader{Reason: fmt.Sprintf("unsupported version %d", version)}
	}

	state.objectCountHeader = binary.BigEndian.Uint32(hdr[8:12])
	if state.objectCountHeader == 0 {
		return &ErrInvalidPackHeader{Reason: "zero objects"}
	}

	return nil
}

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
		return 0, &ErrMalformedPackEntry{
			Offset: startOffset,
			Reason: fmt.Sprintf("inflated size mismatch got %d want %d", contentLen, record.declaredSize),
		}
	}

	endOffset := startOffset + uint64(record.headerLen) + consumedInput
	if endOffset > state.stream.consumed {
		return 0, &ErrMalformedPackEntry{
			Offset: startOffset,
			Reason: fmt.Sprintf("entry end offset overflow got %d > stream %d", endOffset, state.stream.consumed),
		}
	}

	record.packedLen = endOffset - startOffset
	record.dataOffset = startOffset + uint64(record.headerLen)
	if record.packedLen < uint64(record.headerLen) {
		return 0, &ErrMalformedPackEntry{Offset: startOffset, Reason: "negative payload span"}
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
	}

	return endOffset, nil
}

// parseEntryPrefix parses one entry prefix from stream.
func parseEntryPrefix(state *ingestState, startOffset uint64) (objectRecord, error) {
	var record objectRecord
	record.offset = startOffset

	first, err := state.stream.ReadByte()
	if err != nil {
		return record, &ErrMalformedPackEntry{Offset: startOffset, Reason: fmt.Sprintf("read first header byte: %v", err)}
	}

	record.packedType = objecttype.Type((first >> 4) & 0x07)
	size := int64(first & 0x0f)
	headerLen := uint32(1)
	shift := uint(4)
	b := first

	for b&0x80 != 0 {
		b, err = state.stream.ReadByte()
		if err != nil {
			return record, &ErrMalformedPackEntry{Offset: startOffset, Reason: fmt.Sprintf("read size continuation: %v", err)}
		}
		headerLen++
		size |= int64(b&0x7f) << shift
		shift += 7
	}
	if size < 0 {
		return record, &ErrMalformedPackEntry{Offset: startOffset, Reason: "negative declared size"}
	}
	record.declaredSize = size

	switch record.packedType {
	case objecttype.TypeCommit, objecttype.TypeTree, objecttype.TypeBlob, objecttype.TypeTag:
	case objecttype.TypeRefDelta:
		baseRaw := make([]byte, state.algo.Size())
		if err := state.stream.readFull(baseRaw); err != nil {
			return record, &ErrMalformedPackEntry{Offset: startOffset, Reason: fmt.Sprintf("read ref base: %v", err)}
		}
		baseID, err := objectid.FromBytes(state.algo, baseRaw)
		if err != nil {
			return record, &ErrMalformedPackEntry{Offset: startOffset, Reason: fmt.Sprintf("parse ref base: %v", err)}
		}
		record.baseObject = baseID
		headerLen += uint32(len(baseRaw))
	case objecttype.TypeOfsDelta:
		dist, consumed, err := readOfsDistanceFromStream(state.stream)
		if err != nil {
			return record, &ErrMalformedPackEntry{Offset: startOffset, Reason: err.Error()}
		}
		if startOffset <= dist {
			return record, &ErrMalformedPackEntry{Offset: startOffset, Reason: "ofs base offset out of bounds"}
		}
		record.baseOffset = startOffset - dist
		headerLen += uint32(consumed)
	case objecttype.TypeInvalid, objecttype.TypeFuture:
		return record, &ErrMalformedPackEntry{Offset: startOffset, Reason: fmt.Sprintf("unsupported object type %d", record.packedType)}
	default:
		return record, &ErrMalformedPackEntry{Offset: startOffset, Reason: fmt.Sprintf("unsupported object type %d", record.packedType)}
	}

	record.headerLen = headerLen

	return record, nil
}

// drainEntryPayload inflates one entry payload from stream and returns
// (inflatedLength, consumedInput, oidForBaseEntry).
func drainEntryPayload(state *ingestState, record objectRecord) (int64, uint64, objectid.ObjectID, error) {
	var zero objectid.ObjectID
	reader, err := zlib.NewReader(state.stream)
	if err != nil {
		return 0, 0, zero, &ErrMalformedPackEntry{Offset: record.offset, Reason: fmt.Sprintf("open zlib stream: %v", err)}
	}
	defer func() { _ = reader.Close() }()

	var total int64
	if packfmt.IsBaseObjectType(record.packedType) {
		header, ok := objectheader.Encode(record.packedType, record.declaredSize)
		if !ok {
			return 0, 0, zero, &ErrMalformedPackEntry{Offset: record.offset, Reason: "encode object header"}
		}

		hashImpl, err := state.algo.New()
		if err != nil {
			return 0, 0, zero, err
		}
		_, _ = hashImpl.Write(header)

		n, err := io.Copy(hashImpl, reader)
		if err != nil {
			return 0, 0, zero, &ErrMalformedPackEntry{Offset: record.offset, Reason: fmt.Sprintf("inflate base object: %v", err)}
		}
		total = n

		oid, err := objectid.FromBytes(state.algo, hashImpl.Sum(nil))
		if err != nil {
			return 0, 0, zero, err
		}

		return total, reader.InputConsumed(), oid, nil
	}

	if record.packedType == objecttype.TypeOfsDelta || record.packedType == objecttype.TypeRefDelta {
		n, err := io.Copy(io.Discard, reader)
		if err != nil {
			return 0, 0, zero, &ErrMalformedPackEntry{Offset: record.offset, Reason: fmt.Sprintf("inflate delta payload: %v", err)}
		}
		total = n

		return total, reader.InputConsumed(), zero, nil
	}

	return 0, 0, zero, &ErrMalformedPackEntry{Offset: record.offset, Reason: "unsupported payload type"}
}

// readOfsDistanceFromStream reads one ofs-delta encoded distance.
func readOfsDistanceFromStream(reader io.ByteReader) (uint64, int, error) {
	first, err := reader.ReadByte()
	if err != nil {
		return 0, 0, fmt.Errorf("read ofs distance first byte: %w", err)
	}

	dist := uint64(first & 0x7f)
	consumed := 1
	b := first
	for b&0x80 != 0 {
		b, err = reader.ReadByte()
		if err != nil {
			return 0, 0, fmt.Errorf("read ofs distance continuation: %w", err)
		}
		consumed++
		dist = ((dist + 1) << 7) + uint64(b&0x7f)
	}

	return dist, consumed, nil
}

// finalizeStreamPackHash consumes trailer bytes and verifies stream integrity.
// readDeltaHeaderSizes reads source and destination sizes from one delta payload.
func readDeltaHeaderSizes(payload []byte) (int, int, error) {
	reader := &byteSliceReader{data: payload}

	return deltaapply.ReadHeaderSizes(reader)
}

// byteSliceReader implements io.ByteReader on []byte.
type byteSliceReader struct {
	data []byte
	pos  int
}

// ReadByte reads one byte from receiver.
func (reader *byteSliceReader) ReadByte() (byte, error) {
	if reader.pos >= len(reader.data) {
		return 0, io.EOF
	}

	b := reader.data[reader.pos]
	reader.pos++

	return b, nil
}
