package pack

import (
	"fmt"

	"codeberg.org/lindenii/furgit/objecttype"
)

// EntryHeader is one parsed pack entry header prefix.
type EntryHeader struct {
	// Type is the entry type tag from the first header byte.
	Type objecttype.Type
	// Size is the declared resulting object size.
	Size int64
	// HeaderSize is the number of bytes consumed by the type/size header.
	HeaderSize int
}

// ParseEntryHeader parses one pack entry type/size header from data.
func ParseEntryHeader(data []byte) (EntryHeader, error) {
	var zero EntryHeader
	if len(data) == 0 {
		return zero, fmt.Errorf("packfile: truncated entry header")
	}

	first := data[0]
	header := EntryHeader{
		Type:       objecttype.Type((first >> 4) & 0x07),
		Size:       int64(first & 0x0f),
		HeaderSize: 1,
	}

	shift := uint(4)

	b := first
	for b&0x80 != 0 {
		if header.HeaderSize >= len(data) {
			return zero, fmt.Errorf("packfile: truncated entry header")
		}

		b = data[header.HeaderSize]
		header.HeaderSize++
		header.Size |= int64(b&0x7f) << shift
		shift += 7
	}

	if header.Size < 0 {
		return zero, fmt.Errorf("packfile: negative entry size")
	}

	return header, nil
}

// Entry is one parsed pack entry prefix, including any delta base reference
// data that appears before the compressed payload.
type Entry struct {
	// Type is the pack entry type.
	Type objecttype.Type
	// Size is the declared resulting object size.
	Size int64
	// DataOffset is the byte offset from the start of the entry to the zlib
	// payload bytes.
	DataOffset int
	// RefBaseID is the referenced base object ID bytes for ref-delta entries.
	RefBaseID []byte
	// OfsBaseDistance is the backward distance for ofs-delta entries.
	OfsBaseDistance uint64
}

// ParseEntry parses one full pack entry prefix from data.
//
// hashSize must match the hash algorithm size used by the pack/index.
func ParseEntry(data []byte, hashSize int) (Entry, error) {
	var zero Entry

	header, err := ParseEntryHeader(data)
	if err != nil {
		return zero, err
	}

	entry := Entry{
		Type:       header.Type,
		Size:       header.Size,
		DataOffset: header.HeaderSize,
	}

	switch entry.Type {
	case objecttype.TypeCommit, objecttype.TypeTree, objecttype.TypeBlob, objecttype.TypeTag:
		// Base object entries have no extra prefix fields.
	case objecttype.TypeRefDelta:
		if hashSize <= 0 {
			return zero, fmt.Errorf("packfile: invalid hash size %d", hashSize)
		}

		end := entry.DataOffset + hashSize
		if end > len(data) {
			return zero, fmt.Errorf("packfile: truncated ref-delta base id")
		}

		entry.RefBaseID = data[entry.DataOffset:end]
		entry.DataOffset = end
	case objecttype.TypeOfsDelta:
		dist, consumed, err := ParseOfsDeltaDistance(data[entry.DataOffset:])
		if err != nil {
			return zero, err
		}

		entry.OfsBaseDistance = dist
		entry.DataOffset += consumed
	case objecttype.TypeInvalid, objecttype.TypeFuture:
		return zero, fmt.Errorf("packfile: unsupported object type %d", entry.Type)
	default:
		return zero, fmt.Errorf("packfile: unsupported object type %d", entry.Type)
	}

	if entry.DataOffset > len(data) {
		return zero, fmt.Errorf("packfile: entry data offset out of bounds")
	}

	return entry, nil
}
