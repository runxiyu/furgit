package ingest

import (
	"lindenii.org/go/furgit/internal/format/packfile"
	"lindenii.org/go/furgit/object/id"
)

// record is the scanned metadata for one packed entry,
// completed in place as deltas are resolved.
//
// Records are appended in pack-offset order,
// so a record's index in the slice is also its pack order.
type record struct {
	// offset is the entry's start offset in the pack.
	offset uint64

	// headerLen is the entry header length in bytes,
	// so the zlib payload begins at offset+headerLen.
	headerLen uint64

	// packedLen is the total on-disk entry length in bytes,
	// covering the header and the compressed payload.
	packedLen uint64

	// crc32 is the CRC32 of the entry's packed bytes.
	crc32 uint32

	// packedType is the entry type as encoded in the pack.
	packedType packfile.EntryType

	// declaredSize is the declared inflated payload size.
	declaredSize uint64

	// baseOffset is the base entry offset for an ofs-delta.
	baseOffset uint64

	// baseOID is the base object ID for a ref-delta.
	baseOID id.ObjectID

	// objectType is the resolved object type,
	// meaningful once resolved is true.
	objectType packfile.EntryType

	// oid is the resolved object ID,
	// meaningful once resolved is true.
	oid id.ObjectID

	// resolved reports whether oid and objectType are final.
	resolved bool
}

// dataOffset returns the entry's compressed payload start offset.
func (record *record) dataOffset() uint64 {
	return record.offset + record.headerLen
}
