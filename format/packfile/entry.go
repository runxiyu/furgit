package packfile

import (
	"fmt"

	objecttype "codeberg.org/lindenii/furgit/object/type"
)

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
