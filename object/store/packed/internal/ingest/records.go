package ingest

import (
	objectid "lindenii.org/go/furgit/object/id"
	objecttype "lindenii.org/go/furgit/object/type"
)

// objectRecord stores metadata for one packed object entry.
type objectRecord struct {
	// offset is the entry start offset in the pack file.
	offset uint64
	// headerLen is packed entry header length in bytes.
	headerLen uint32
	// packedLen is total packed entry length in bytes.
	packedLen uint64
	// crc32 is the CRC over the full packed entry.
	crc32 uint32
	// packedType is the entry type tag from the pack stream.
	packedType objecttype.Type
	// realType is canonical object type after delta resolution.
	realType objecttype.Type
	// declaredSize is the declared output object size for this entry.
	declaredSize int64
	// dataOffset is compressed payload start offset for this entry.
	dataOffset uint64
	// baseOffset is OFS base offset when packedType is OFS delta.
	baseOffset uint64
	// baseObject is REF base object ID when packedType is REF delta.
	baseObject objectid.ObjectID
	// objectID is final resolved object ID.
	objectID objectid.ObjectID
	// resolved reports whether objectID/realType are finalized.
	resolved bool
}

// ofsDeltaRef maps one OFS delta record to its base offset.
type ofsDeltaRef struct {
	baseOffset uint64
	recordIdx  int
}

// refDeltaRef maps one REF delta record to its base object ID.
type refDeltaRef struct {
	baseObject objectid.ObjectID
	recordIdx  int
}
