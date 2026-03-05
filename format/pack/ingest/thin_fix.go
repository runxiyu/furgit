package ingest

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"os"

	"codeberg.org/lindenii/furgit/internal/compress/zlib"
	"codeberg.org/lindenii/furgit/internal/intconv"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
)

// maybeFixThin appends missing bases and rewrites pack header/trailer when needed.
func maybeFixThin(state *ingestState) error {
	if len(state.unresolvedRefDeltas) == 0 {
		return nil
	}

	if !state.fixThin {
		return &ErrThinPackUnresolved{Count: len(state.unresolvedRefDeltas)}
	}

	if state.base == nil {
		return &ErrThinPackUnresolved{Count: len(state.unresolvedRefDeltas)}
	}

	hashSize := int64(state.algo.Size())

	info, err := state.packFile.Stat()
	if err != nil {
		return err
	}

	size := info.Size()
	if size < hashSize {
		return fmt.Errorf("format/pack/ingest: pack too short to trim trailer")
	}

	newEnd := size - hashSize

	err = state.packFile.Truncate(newEnd)
	if err != nil {
		return err
	}

	consumed, err := intconv.Int64ToUint64(newEnd)
	if err != nil {
		return err
	}

	state.stream.consumed = consumed

	baseIDs := unresolvedThinBaseIDs(state)
	for _, id := range baseIDs {
		ty, content, err := state.base.ReadBytesContent(id)
		if err != nil {
			continue
		}

		_, err = appendBaseObject(state, id, ty, content)
		if err != nil {
			return err
		}

		state.thinFixed = true
	}

	err = rewritePackHeaderAndTrailer(state)
	if err != nil {
		return err
	}

	return nil
}

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
	_, _ = crc.Write(header)
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

// fileSectionWriter writes sequentially to file via WriteAt at one base offset.
type fileSectionWriter struct {
	file *os.File
	off  int64
	pos  int64
}

// Write writes src at current section position.
func (writer *fileSectionWriter) Write(src []byte) (int, error) {
	if len(src) == 0 {
		return 0, nil
	}

	n, err := writer.file.WriteAt(src, writer.off+writer.pos)
	writer.pos += int64(n)

	return n, err
}

// countingWriter counts bytes written to dst.
type countingWriter struct {
	dst io.Writer
	n   int
}

// Write writes src to dst and tracks output byte count.
func (writer *countingWriter) Write(src []byte) (int, error) {
	n, err := writer.dst.Write(src)
	writer.n += n

	return n, err
}

// rewritePackHeaderAndTrailer rewrites object count and trailer hash using ReadAt/WriteAt.
func rewritePackHeaderAndTrailer(state *ingestState) error {
	var countRaw [4]byte

	recordCountUint32, err := intconv.IntToUint32(len(state.records))
	if err != nil {
		return err
	}

	binary.BigEndian.PutUint32(countRaw[:], recordCountUint32)

	_, err = state.packFile.WriteAt(countRaw[:], 8)
	if err != nil {
		return err
	}

	info, err := state.packFile.Stat()
	if err != nil {
		return err
	}

	endWithoutTrailer := info.Size()

	hashImpl, err := state.algo.New()
	if err != nil {
		return err
	}

	var (
		buf [128 << 10]byte
		pos int64
	)
	for pos < endWithoutTrailer {
		want := int64(len(buf))

		remaining := endWithoutTrailer - pos
		if remaining < want {
			want = remaining
		}

		n, err := state.packFile.ReadAt(buf[:want], pos)
		if err != nil && err != io.EOF {
			return err
		}

		if n == 0 {
			return io.ErrUnexpectedEOF
		}

		_, _ = hashImpl.Write(buf[:n])
		pos += int64(n)
	}

	sum := hashImpl.Sum(nil)

	_, err = state.packFile.WriteAt(sum, endWithoutTrailer)
	if err != nil {
		return err
	}

	packHash, err := objectid.FromBytes(state.algo, sum)
	if err != nil {
		return err
	}

	state.packHash = packHash
	state.objectCountHeader = recordCountUint32

	sumLenInt64 := int64(len(sum))

	newConsumed, err := intconv.Int64ToUint64(endWithoutTrailer + sumLenInt64)
	if err != nil {
		return err
	}

	state.stream.consumed = newConsumed

	return nil
}

// encodePackEntryHeader encodes one non-delta packed entry header.
func encodePackEntryHeader(ty objecttype.Type, size int64) []byte {
	var out [16]byte

	n := 0

	s, err := intconv.Int64ToUint64(size)
	if err != nil {
		panic(err)
	}

	c := byte((uint8(ty) << 4) | byte(s&0x0f))

	s >>= 4
	for s != 0 {
		out[n] = c | 0x80
		n++
		c = byte(s & 0x7f)
		s >>= 7
	}

	out[n] = c
	n++

	return append([]byte(nil), out[:n]...)
}
