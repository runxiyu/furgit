package packfile

import (
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	// signature is the 4-byte "PACK" magic at the start of pack files.
	signature = 0x5041434b

	// version is the only pack version this package reads or writes.
	version = 2
)

// HeaderLen is the size of a pack header:
// signature, version, and object count.
const HeaderLen = 12

// ErrMalformedHeader reports that
// a pack header is truncated,
// or has a bad signature or unsupported version.
var ErrMalformedHeader = errors.New("internal/format/packfile: malformed pack header")

// Header is one parsed pack header.
type Header struct {
	// ObjectCount is the number of object entries
	// declared to follow the header.
	ObjectCount uint32
}

// ParseHeader parses and validates the pack header
// at the start of data.
func ParseHeader(data []byte) (Header, error) {
	var zero Header

	if len(data) < HeaderLen {
		return zero, fmt.Errorf("%w: truncated", ErrMalformedHeader)
	}

	if binary.BigEndian.Uint32(data[0:4]) != signature {
		return zero, fmt.Errorf("%w: bad signature", ErrMalformedHeader)
	}

	parsedVersion := binary.BigEndian.Uint32(data[4:8])
	if parsedVersion != version {
		return zero, fmt.Errorf("%w: unsupported version %d", ErrMalformedHeader, parsedVersion)
	}

	return Header{
		ObjectCount: binary.BigEndian.Uint32(data[8:12]),
	}, nil
}

// AppendHeader appends an encoded pack header for objectCount to dst.
func AppendHeader(dst []byte, objectCount uint32) []byte {
	dst = binary.BigEndian.AppendUint32(dst, signature)
	dst = binary.BigEndian.AppendUint32(dst, version)
	dst = binary.BigEndian.AppendUint32(dst, objectCount)

	return dst
}
