package packfile

import (
	"errors"
	"fmt"

	"lindenii.org/go/furgit/object/id"
)

// ErrMalformedEntryHeader reports that
// a packfile entry header is truncated, overlong,
// has an unsupported entry type,
// or declares a size that overflows uint64.
var ErrMalformedEntryHeader = errors.New("internal/format/packfile: malformed entry header")

// MaxTypeSizeLen is the maximum encoded length
// of the type/size prefix of an entry header.
// Every uint64 size is encodable within this bound,
// and [ParseEntryHeader] rejects longer prefixes.
const MaxTypeSizeLen = 10

// MaxEntryHeaderLen returns the maximum encoded length
// of a full entry header
// for packs whose object IDs are hashSize bytes.
//
// Callers parsing from a stream may buffer
// MaxEntryHeaderLen bytes
// (or fewer if the pack data ends sooner)
// and parse with [ParseEntryHeader];
// no valid entry header is longer.
func MaxEntryHeaderLen(hashSize int) int {
	return MaxTypeSizeLen + max(hashSize, MaxOfsDeltaDistanceLen)
}

// EntryHeader is one parsed packfile entry header:
// everything from the start of the entry
// up to its zlib payload.
type EntryHeader struct {
	// Type is the packfile entry type.
	Type EntryType

	// Size is the declared inflated size
	// of the entry's payload.
	// For delta entries this is the delta size,
	// not the reconstructed object size.
	Size uint64

	// HeaderLen is the number of bytes
	// the header occupies in the pack;
	// the zlib payload begins HeaderLen bytes
	// after the start of the entry.
	HeaderLen int

	// RefBase holds the base object ID
	// for ref-delta entries.
	// Only the first hashSize bytes are meaningful.
	RefBase [id.MaxObjectIDSize]byte

	// OfsDistance is the backward distance
	// from the start of this entry
	// to the start of the base entry,
	// for ofs-delta entries.
	OfsDistance uint64
}

// ParseEntryHeader parses one packfile entry header
// from the beginning of data.
//
// hashSize must be the object ID size
// of the pack's object format.
//
// data need not contain the whole entry;
// [MaxEntryHeaderLen] bytes always suffice.
// Headers of types [EntryTypeInvalid] and [EntryTypeFuture]
// are rejected as malformed.
func ParseEntryHeader(data []byte, hashSize int) (EntryHeader, error) {
	var zero EntryHeader

	if hashSize <= 0 || hashSize > id.MaxObjectIDSize {
		return zero, fmt.Errorf("internal/format/packfile: invalid hash size %d", hashSize)
	}

	if len(data) == 0 {
		return zero, fmt.Errorf("%w: truncated type/size prefix", ErrMalformedEntryHeader)
	}

	first := data[0]
	header := EntryHeader{
		Type:      EntryType((first >> 4) & 0x07),
		Size:      uint64(first & 0x0f),
		HeaderLen: 1,
	}

	shift := uint(4)

	b := first
	for b&0x80 != 0 {
		if header.HeaderLen >= MaxTypeSizeLen {
			return zero, fmt.Errorf("%w: overlong type/size prefix", ErrMalformedEntryHeader)
		}

		if header.HeaderLen >= len(data) {
			return zero, fmt.Errorf("%w: truncated type/size prefix", ErrMalformedEntryHeader)
		}

		b = data[header.HeaderLen]
		header.HeaderLen++

		group := uint64(b & 0x7f)
		if group<<shift>>shift != group {
			return zero, fmt.Errorf("%w: size overflows uint64", ErrMalformedEntryHeader)
		}

		header.Size |= group << shift
		shift += 7
	}

	switch header.Type {
	case EntryTypeCommit, EntryTypeTree, EntryTypeBlob, EntryTypeTag:
		// Base entries have nothing between the type/size prefix and the payload.
	case EntryTypeRefDelta:
		end := header.HeaderLen + hashSize
		if end > len(data) {
			return zero, fmt.Errorf("%w: truncated ref-delta base ID", ErrMalformedEntryHeader)
		}

		copy(header.RefBase[:], data[header.HeaderLen:end])
		header.HeaderLen = end
	case EntryTypeOfsDelta:
		dist, consumed, err := ParseOfsDeltaDistance(data[header.HeaderLen:])
		if err != nil {
			return zero, fmt.Errorf("%w: %w", ErrMalformedEntryHeader, err)
		}

		header.OfsDistance = dist
		header.HeaderLen += consumed
	case EntryTypeInvalid, EntryTypeFuture:
		return zero, fmt.Errorf("%w: unsupported entry type", ErrMalformedEntryHeader)
	default:
		return zero, fmt.Errorf("%w: unsupported entry type", ErrMalformedEntryHeader)
	}

	return header, nil
}

// AppendTypeSize appends the type/size prefix encoding
// of an entry header to dst.
//
// entryType must be a valid on-disk entry type;
// [EntryTypeInvalid] and [EntryTypeFuture] and
// values that do not fit in three bits
// produce garbage encodings.
func AppendTypeSize(dst []byte, entryType EntryType, size uint64) []byte {
	b := byte(entryType)<<4 | byte(size&0x0f)
	size >>= 4

	for size != 0 {
		dst = append(dst, b|0x80)
		b = byte(size & 0x7f)
		size >>= 7
	}

	return append(dst, b)
}
