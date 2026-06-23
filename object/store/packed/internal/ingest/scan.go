package ingest

import (
	"bytes"
	"errors"
	"fmt"
	"hash"
	"hash/crc32"
	"io"

	"lindenii.org/go/furgit/internal/compress/zlib"
	"lindenii.org/go/furgit/internal/format/packfile"
	"lindenii.org/go/furgit/internal/progress"
	"lindenii.org/go/furgit/object/header"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store"
	"lindenii.org/go/lgo/intconv"
)

// scanBufferSize is the stream scanner's fixed input window size.
const scanBufferSize = 64 << 10

// scanner reads one pack stream,
// mirroring consumed bytes into the destination pack file
// while maintaining the running pack hash and a per-entry CRC.
//
// It implements [io.Reader] and [io.ByteReader]
// so a zlib reader can consume an entry payload through it
// without reading past the end of that compressed stream.
type scanner struct {
	src io.Reader
	dst io.Writer

	// buf[off:n] is the unread window.
	buf []byte
	off int
	n   int

	// consumed counts stream bytes consumed so far.
	consumed int

	// hash accumulates the pack hash over consumed bytes
	// while hashing is true.
	hash    hash.Hash
	hashing bool

	// crc accumulates the CRC of the current entry
	// while crcing is true.
	crc    uint32
	crcing bool
}

// newScanner constructs one scanner mirroring src into dst,
// seeding the running hash from the already-consumed pack header.
func newScanner(src io.Reader, dst io.Writer, packHash hash.Hash) *scanner {
	return &scanner{
		src:      src,
		dst:      dst,
		buf:      make([]byte, scanBufferSize),
		consumed: packfile.HeaderLen,
		hash:     packHash,
		hashing:  true,
		crc:      0,
		crcing:   false,
	}
}

// readPackHeader reads and validates the pack header from src,
// returning the raw header and its declared object count.
func readPackHeader(src io.Reader) ([packfile.HeaderLen]byte, int, error) {
	var raw [packfile.HeaderLen]byte

	_, err := io.ReadFull(src, raw[:])
	if err != nil {
		return raw, 0, fmt.Errorf("%w: reading header: %w", ErrMalformedPack, err)
	}

	packHeader, err := packfile.ParseHeader(raw[:])
	if err != nil {
		return raw, 0, fmt.Errorf("%w: %w", ErrMalformedPack, err)
	}

	count, err := intconv.Uint32ToInt(packHeader.ObjectCount)
	if err != nil {
		return raw, 0, fmt.Errorf("%w: object count: %w", ErrMalformedPack, err)
	}

	return raw, count, nil
}

// Read implements [io.Reader].
func (scanner *scanner) Read(dst []byte) (int, error) {
	if len(dst) == 0 {
		return 0, nil
	}

	err := scanner.ensureAvailable()
	if err != nil {
		return 0, err
	}

	read := min(len(dst), scanner.n-scanner.off)

	copy(dst, scanner.buf[scanner.off:scanner.off+read])

	err = scanner.use(read)
	if err != nil {
		return 0, err
	}

	return read, nil
}

// ReadByte implements [io.ByteReader] without allocation.
func (scanner *scanner) ReadByte() (byte, error) {
	err := scanner.ensureAvailable()
	if err != nil {
		return 0, err
	}

	b := scanner.buf[scanner.off]

	err = scanner.use(1)
	if err != nil {
		return 0, err
	}

	return b, nil
}

// ensureAvailable makes at least one unread byte available,
// returning [io.EOF] once the source is exhausted.
func (scanner *scanner) ensureAvailable() error {
	for scanner.n-scanner.off == 0 {
		err := scanner.flushPrefix()
		if err != nil {
			return err
		}

		read, err := scanner.src.Read(scanner.buf[scanner.n:])
		scanner.n += read

		if err != nil {
			if errors.Is(err, io.EOF) {
				if scanner.n-scanner.off == 0 {
					return io.EOF
				}

				return nil
			}

			return fmt.Errorf("object/store/packed/internal/ingest: reading pack: %w", err)
		}

		if read == 0 && scanner.n-scanner.off == 0 {
			return io.ErrNoProgress
		}
	}

	return nil
}

// peekHeader returns the unread window grown to at most maxLen bytes
// without consuming, tolerating an early end of stream.
func (scanner *scanner) peekHeader(maxLen int) ([]byte, error) {
	maxLen = min(maxLen, len(scanner.buf))

	for scanner.n-scanner.off < maxLen {
		err := scanner.flushPrefix()
		if err != nil {
			return nil, err
		}

		read, err := scanner.src.Read(scanner.buf[scanner.n:])
		scanner.n += read

		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return nil, fmt.Errorf("object/store/packed/internal/ingest: reading pack: %w", err)
		}

		if read == 0 {
			break
		}
	}

	if scanner.n-scanner.off == 0 {
		return nil, fmt.Errorf("%w: unexpected end of stream", ErrMalformedPack)
	}

	return scanner.buf[scanner.off:scanner.n], nil
}

// use consumes n unread bytes,
// folding them into the running hash and entry CRC as enabled.
func (scanner *scanner) use(n int) error {
	chunk := scanner.buf[scanner.off : scanner.off+n]

	if scanner.hashing {
		_, err := scanner.hash.Write(chunk)
		if err != nil {
			return fmt.Errorf("object/store/packed/internal/ingest: hashing pack: %w", err)
		}
	}

	if scanner.crcing {
		scanner.crc = crc32.Update(scanner.crc, crc32.IEEETable, chunk)
	}

	scanner.off += n
	scanner.consumed += n

	return nil
}

// flushPrefix writes the consumed buffer prefix to the destination
// and compacts the unread window to the start of the buffer.
func (scanner *scanner) flushPrefix() error {
	if scanner.off == 0 {
		return nil
	}

	_, err := scanner.dst.Write(scanner.buf[:scanner.off])
	if err != nil {
		return fmt.Errorf("object/store/packed/internal/ingest: writing pack: %w", err)
	}

	unread := scanner.n - scanner.off

	copy(scanner.buf, scanner.buf[scanner.off:scanner.n])

	scanner.off = 0
	scanner.n = unread

	return nil
}

// beginCRC starts CRC accumulation for one entry.
func (scanner *scanner) beginCRC() {
	scanner.crc = 0
	scanner.crcing = true
}

// endCRC ends CRC accumulation and returns the entry CRC.
func (scanner *scanner) endCRC() uint32 {
	crc := scanner.crc
	scanner.crc = 0
	scanner.crcing = false

	return crc
}

// finishTrailer reads and verifies the pack trailer hash,
// flushing the remaining buffered pack bytes to the destination.
//
// The trailer is mirrored to the destination but excluded from the pack hash.
func (scanner *scanner) finishTrailer(hashSize int) ([]byte, error) {
	trailer := make([]byte, hashSize)

	scanner.hashing = false

	_, err := io.ReadFull(scanner, trailer)
	if err != nil {
		return nil, fmt.Errorf("%w: reading trailer: %w", ErrMalformedPack, err)
	}

	if scanner.n-scanner.off > 0 {
		return nil, fmt.Errorf("%w: trailing data after pack", ErrMalformedPack)
	}

	if !bytes.Equal(scanner.hash.Sum(nil), trailer) {
		return nil, fmt.Errorf("%w: trailer hash mismatch", ErrMalformedPack)
	}

	err = scanner.flushPrefix()
	if err != nil {
		return nil, err
	}

	return trailer, nil
}

// streamAndScan streams the pack body to the temporary pack file,
// scanning one record per declared object and verifying the trailer.
func (ingestion *ingestion) streamAndScan() error {
	meter := progress.New(progress.Options{
		Writer:     ingestion.opts.Progress,
		Title:      "receiving objects",
		Total:      ingestion.headerCount,
		Delay:      0,
		Sparse:     false,
		Throughput: true,
	})

	for done := range ingestion.headerCount {
		err := ingestion.ctx.Err()
		if err != nil {
			return fmt.Errorf("object/store/packed/internal/ingest: %w", err)
		}

		err = ingestion.scanEntry(ingestion.scanner.consumed)
		if err != nil {
			return err
		}

		meter.Set(done+1, ingestion.scanner.consumed)
	}

	meter.Stop("done")

	trailer, err := ingestion.scanner.finishTrailer(ingestion.objectFormat.Size())
	if err != nil {
		return err
	}

	packHash, err := ingestion.objectFormat.FromBytes(trailer)
	if err != nil {
		return fmt.Errorf("object/store/packed/internal/ingest: %w", err)
	}

	ingestion.packHash = packHash

	return nil
}

// scanEntry scans the entry beginning at start into one record.
func (ingestion *ingestion) scanEntry(start int) error {
	ingestion.scanner.beginCRC()

	rec, err := ingestion.scanHeader(start)
	if err != nil {
		return err
	}

	inflated, oid, err := ingestion.drainPayload(&rec)
	if err != nil {
		return err
	}

	if inflated != int64(rec.declaredSize) {
		return fmt.Errorf(
			"%w: entry at %d: inflated size %d differs from declared %d",
			ErrMalformedPack, start, inflated, rec.declaredSize,
		)
	}

	rec.packedLen = ingestion.scanner.consumed - start
	rec.crc32 = ingestion.scanner.endCRC()

	if rec.packedType.IsBase() {
		rec.oid = oid
		rec.resolved = true
	} else {
		ingestion.deltaCount++
	}

	index := len(ingestion.records)
	ingestion.records = append(ingestion.records, rec)
	ingestion.byOffset[rec.offset] = index

	if rec.resolved {
		ingestion.byOID[rec.oid] = index
	}

	return nil
}

// scanHeader parses and consumes the entry header at start.
func (ingestion *ingestion) scanHeader(start int) (record, error) {
	var rec record

	rec.offset = start

	window, err := ingestion.scanner.peekHeader(packfile.MaxEntryHeaderLen(ingestion.objectFormat.Size()))
	if err != nil {
		return rec, err
	}

	entryHeader, err := packfile.ParseEntryHeader(window, ingestion.objectFormat.Size())
	if err != nil {
		return rec, fmt.Errorf("%w: entry at %d: %w", ErrMalformedPack, start, err)
	}

	declaredSize, err := intconv.Uint64ToInt(entryHeader.Size)
	if err != nil {
		return rec, fmt.Errorf("%w: entry at %d: declared size overflows int: %w", ErrMalformedPack, start, err)
	}

	limit := ingestion.opts.MaxObjectSize
	if limit > 0 && declaredSize > limit {
		return rec, fmt.Errorf("%w: entry at %d: declared size %d exceeds limit %d", store.ErrObjectTooLarge, start, declaredSize, limit)
	}

	rec.packedType = entryHeader.Type
	rec.declaredSize = declaredSize
	rec.headerLen = entryHeader.HeaderLen

	switch entryHeader.Type {
	case packfile.EntryTypeCommit, packfile.EntryTypeTree, packfile.EntryTypeBlob, packfile.EntryTypeTag:
	case packfile.EntryTypeOfsDelta:
		dist, err := intconv.Uint64ToInt(entryHeader.OfsDistance)
		if err != nil || dist == 0 || dist > start {
			return rec, fmt.Errorf("%w: entry at %d: ofs-delta base out of bounds", ErrMalformedPack, start)
		}

		rec.baseOffset = start - dist
	case packfile.EntryTypeRefDelta:
		baseID, err := ingestion.objectFormat.FromBytes(entryHeader.RefBase[:ingestion.objectFormat.Size()])
		if err != nil {
			return rec, fmt.Errorf("object/store/packed/internal/ingest: %w", err)
		}

		rec.baseOID = baseID
	case packfile.EntryTypeInvalid, packfile.EntryTypeFuture:
		return rec, fmt.Errorf("%w: entry at %d: unsupported entry type", ErrMalformedPack, start)
	}

	err = ingestion.scanner.use(entryHeader.HeaderLen)
	if err != nil {
		return rec, err
	}

	return rec, nil
}

// drainPayload consumes one entry's compressed payload from the stream,
// returning its inflated length and, for base entries, its object ID.
func (ingestion *ingestion) drainPayload(rec *record) (int64, id.ObjectID, error) {
	var zero id.ObjectID

	zr, err := zlib.NewReader(ingestion.scanner)
	if err != nil {
		return 0, zero, fmt.Errorf("%w: entry at %d: %w", ErrMalformedPack, rec.offset, err)
	}

	defer func() { _ = zr.Close() }()

	if !rec.packedType.IsBase() {
		read, err := io.Copy(io.Discard, zr)
		if err != nil {
			return 0, zero, fmt.Errorf("%w: entry at %d: %w", ErrMalformedPack, rec.offset, err)
		}

		return read, zero, nil
	}

	objectType, err := rec.packedType.ObjectType()
	if err != nil {
		return 0, zero, fmt.Errorf("object/store/packed/internal/ingest: %w", err)
	}

	hashImpl, err := ingestion.objectFormat.New()
	if err != nil {
		return 0, zero, fmt.Errorf("object/store/packed/internal/ingest: %w", err)
	}

	_, _ = hashImpl.Write(header.Append(nil, objectType, rec.declaredSize))

	read, err := io.Copy(hashImpl, zr)
	if err != nil {
		return 0, zero, fmt.Errorf("%w: entry at %d: %w", ErrMalformedPack, rec.offset, err)
	}

	oid, err := ingestion.objectFormat.FromBytes(hashImpl.Sum(nil))
	if err != nil {
		return 0, zero, fmt.Errorf("object/store/packed/internal/ingest: %w", err)
	}

	return read, oid, nil
}
