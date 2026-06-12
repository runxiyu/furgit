package packidx

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"

	"lindenii.org/go/furgit/internal/stickyio"
	"lindenii.org/go/furgit/object/id"
)

// ErrInvalidEntries reports that
// entries supplied for an index write
// are unsorted, duplicated, or not representable
// in the pack index format.
var ErrInvalidEntries = errors.New("internal/format/packidx: invalid entries")

// Entry is one object record for an index write.
type Entry struct {
	// OID holds the object ID bytes;
	// only the first hash-size bytes are meaningful.
	OID [id.MaxObjectIDSize]byte
	// Offset is the entry's pack file offset.
	Offset uint64
	// CRC32 is the CRC32 of the entry's packed data.
	CRC32 uint32
}

// Write writes one pack index over entries to w.
//
// entries must be sorted by object ID without duplicates.
// packHash must be the pack's trailer hash;
// Write panics when its length does not match the object format.
func Write(w io.Writer, objectFormat id.ObjectFormat, entries []Entry, packHash []byte) error {
	hashSize := objectFormat.Size()
	if hashSize == 0 {
		return id.ErrInvalidObjectFormat
	}

	if len(packHash) != hashSize {
		panic("internal/format/packidx: invalid pack hash length")
	}

	if len(entries) > math.MaxUint32 {
		return fmt.Errorf("%w: too many entries", ErrInvalidEntries)
	}

	for i := 1; i < len(entries); i++ {
		if bytes.Compare(entries[i-1].OID[:hashSize], entries[i].OID[:hashSize]) >= 0 {
			return fmt.Errorf("%w: not sorted by object ID", ErrInvalidEntries)
		}
	}

	hashImpl, err := objectFormat.New()
	if err != nil {
		return fmt.Errorf("internal/format/packidx: %w", err)
	}

	bw := bufio.NewWriter(io.MultiWriter(w, hashImpl))
	sw := stickyio.New(bw)

	sw.PutUint32(signature)
	sw.PutUint32(version)

	var counts [256]uint32
	for i := range entries {
		counts[entries[i].OID[0]]++
	}

	cumulative := uint32(0)
	for _, count := range counts {
		cumulative += count
		sw.PutUint32(cumulative)
	}

	for i := range entries {
		sw.Put(entries[i].OID[:hashSize])
	}

	for i := range entries {
		sw.PutUint32(entries[i].CRC32)
	}

	var largeOffsets []uint64

	for i := range entries {
		offset := entries[i].Offset
		if offset < largeOffsetFlag {
			sw.PutUint32(uint32(offset))

			continue
		}

		slot := len(largeOffsets)
		if slot >= largeOffsetFlag {
			return fmt.Errorf("%w: too many large offsets", ErrInvalidEntries)
		}

		sw.PutUint32(largeOffsetFlag | uint32(slot))

		largeOffsets = append(largeOffsets, offset)
	}

	for _, offset := range largeOffsets {
		sw.PutUint64(offset)
	}

	sw.Put(packHash)

	err = sw.Err()
	if err != nil {
		return fmt.Errorf("internal/format/packidx: %w", err)
	}

	err = bw.Flush()
	if err != nil {
		return fmt.Errorf("internal/format/packidx: %w", err)
	}

	_, err = w.Write(hashImpl.Sum(nil))
	if err != nil {
		return fmt.Errorf("internal/format/packidx: %w", err)
	}

	return nil
}
