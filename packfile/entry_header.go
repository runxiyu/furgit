package packfile

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
