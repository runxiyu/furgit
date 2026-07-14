package ingest

import (
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"

	"lindenii.org/go/furgit/internal/compress/zlib"
	"lindenii.org/go/furgit/internal/format/packfile"
	"lindenii.org/go/furgit/internal/progress"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store"
	"lindenii.org/go/furgit/object/typ"
	"lindenii.org/go/lgo/intconv"
)

// fixThin completes a thin pack
// by appending the external bases it references,
// rewriting the pack header and trailer,
// and resolving the deltas reached from the appended bases.
func (ingestion *ingestion) fixThin(external []id.ObjectID, adjacency adjacency, meter *progress.Meter) error {
	if ingestion.opts.ThinBase == nil {
		return ErrThinPackNotPermitted
	}

	hashSize := ingestion.objectFormat.Size()
	if ingestion.scanner.consumed < packfile.HeaderLen+hashSize {
		return fmt.Errorf("%w: pack shorter than trailer", ErrMalformedPack)
	}

	// Drop the trailer from the write cursor.
	ingestion.scanner.consumed -= hashSize

	appended := make([]int, 0, len(external))

	for _, baseOID := range external {
		err := ingestion.ctx.Err()
		if err != nil {
			return fmt.Errorf("object/store/packed/internal/ingest: %w", err)
		}

		ty, content, err := ingestion.opts.ThinBase.ReadBytesContent(baseOID)
		if errors.Is(err, store.ErrObjectNotFound) {
			continue
		}

		if err != nil {
			return fmt.Errorf("object/store/packed/internal/ingest: reading thin base %s: %w", baseOID, err)
		}

		index, err := ingestion.appendBaseObject(baseOID, ty, content)
		if err != nil {
			return err
		}

		appended = append(appended, index)
	}

	err := ingestion.rewriteHeaderTrailer()
	if err != nil {
		return err
	}

	err = ingestion.resolveFrom(appended, adjacency, meter)
	if err != nil {
		return err
	}

	missing := ingestion.unresolvedExternalBases()
	if len(missing) > 0 {
		return &ThinBasesMissingError{OIDs: missing}
	}

	if ingestion.countUnresolved() > 0 {
		return fmt.Errorf("%w: unresolvable delta entries after thin completion", ErrMalformedPack)
	}

	ingestion.thinFixed = len(appended) > 0

	return nil
}

// appendBaseObject appends one external thin base
// as a non-delta pack entry at the current write cursor,
// verifying that its content hashes to the requested object ID.
func (ingestion *ingestion) appendBaseObject(objectID id.ObjectID, objectType typ.Type, content []byte) (int, error) {
	entryType, err := packfile.EntryTypeFromObjectType(objectType)
	if err != nil {
		return 0, fmt.Errorf("object/store/packed/internal/ingest: %w", err)
	}

	computed, err := ingestion.hashObject(objectType, content)
	if err != nil {
		return 0, err
	}

	if computed != objectID {
		return 0, fmt.Errorf("%w: thin base %s content hashes to %s", ErrMalformedPack, objectID, computed)
	}

	start := ingestion.scanner.consumed
	startOffset := int64(start)

	headerBytes := packfile.AppendTypeSize(nil, entryType, uint64(len(content)))

	_, err = ingestion.packFile.WriteAt(headerBytes, startOffset)
	if err != nil {
		return 0, fmt.Errorf("object/store/packed/internal/ingest: writing thin base header: %w", err)
	}

	crc := crc32.NewIEEE()
	_, _ = crc.Write(headerBytes)

	dataOffset := startOffset + int64(len(headerBytes))
	writer := &offsetWriter{file: ingestion.packFile, offset: dataOffset}

	zw := zlib.NewWriter(io.MultiWriter(writer, crc))

	_, err = zw.Write(content)
	if err != nil {
		_ = zw.Close()

		return 0, fmt.Errorf("object/store/packed/internal/ingest: compressing thin base: %w", err)
	}

	err = zw.Close()
	if err != nil {
		return 0, fmt.Errorf("object/store/packed/internal/ingest: compressing thin base: %w", err)
	}

	headerLen := len(headerBytes)
	packedLen := headerLen + writer.written
	ingestion.scanner.consumed = start + packedLen

	rec := record{
		offset:       start,
		headerLen:    headerLen,
		packedLen:    packedLen,
		crc32:        crc.Sum32(),
		packedType:   entryType,
		declaredSize: len(content),
		baseOffset:   0,
		baseOID:      id.ObjectID{},
		oid:          objectID,
		resolved:     true,
	}

	index := len(ingestion.records)
	ingestion.records = append(ingestion.records, rec)
	ingestion.byOffset[start] = index
	ingestion.byOID.Store(objectID, index)

	return index, nil
}

// rewriteHeaderTrailer updates the pack object count
// and recomputes the pack trailer hash
// over the entries left after thin completion.
func (ingestion *ingestion) rewriteHeaderTrailer() error {
	count, err := intconv.IntToUint32(len(ingestion.records))
	if err != nil {
		return fmt.Errorf("object/store/packed/internal/ingest: %w", err)
	}

	_, err = ingestion.packFile.WriteAt(packfile.AppendHeader(nil, count), 0)
	if err != nil {
		return fmt.Errorf("object/store/packed/internal/ingest: rewriting header: %w", err)
	}

	bodyEnd := int64(ingestion.scanner.consumed)

	hashImpl, err := ingestion.objectFormat.New()
	if err != nil {
		return fmt.Errorf("object/store/packed/internal/ingest: %w", err)
	}

	_, err = io.Copy(hashImpl, io.NewSectionReader(ingestion.packFile, 0, bodyEnd))
	if err != nil {
		return fmt.Errorf("object/store/packed/internal/ingest: rehashing pack: %w", err)
	}

	sum := hashImpl.Sum(nil)

	_, err = ingestion.packFile.WriteAt(sum, bodyEnd)
	if err != nil {
		return fmt.Errorf("object/store/packed/internal/ingest: writing trailer: %w", err)
	}

	packHash, err := ingestion.objectFormat.FromBytes(sum)
	if err != nil {
		return fmt.Errorf("object/store/packed/internal/ingest: %w", err)
	}

	ingestion.packHash = packHash

	return nil
}

// offsetWriter writes to a file via WriteAt,
// advancing sequentially from a base offset
// and counting the bytes written.
type offsetWriter struct {
	file    *os.File
	offset  int64
	written int //exhaustruct:optional
}

// Write implements [io.Writer].
func (writer *offsetWriter) Write(p []byte) (int, error) {
	n, err := writer.file.WriteAt(p, writer.offset)
	writer.offset += int64(n)
	writer.written += n

	return n, err //nolint:wrapcheck
}
