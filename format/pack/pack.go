// Package pack provides Git packfile format parsing primitives.
package pack

import (
	"fmt"

	"codeberg.org/lindenii/furgit/objecttype"
)

// Signature is the 4-byte "PACK" magic at the start of pack files.
const Signature = 0x5041434b

// VersionSupported reports whether one pack version is supported.
func VersionSupported(version uint32) bool {
	return version == 2 || version == 3
}

// IsBaseObjectType reports whether ty is one of the four canonical object
// types encoded directly in pack entries.
func IsBaseObjectType(ty objecttype.Type) bool {
	switch ty {
	case objecttype.TypeCommit, objecttype.TypeTree, objecttype.TypeBlob, objecttype.TypeTag:
		return true
	case objecttype.TypeInvalid, objecttype.TypeFuture, objecttype.TypeOfsDelta, objecttype.TypeRefDelta:
		return false
	default:
		return false
	}
}

// ParseOfsDeltaDistance parses one ofs-delta backward distance.
func ParseOfsDeltaDistance(buf []byte) (uint64, int, error) {
	if len(buf) == 0 {
		return 0, 0, fmt.Errorf("format/pack: malformed ofs-delta distance")
	}

	b := buf[0]
	dist := uint64(b & 0x7f)

	consumed := 1
	for b&0x80 != 0 {
		if consumed >= len(buf) {
			return 0, 0, fmt.Errorf("format/pack: malformed ofs-delta distance")
		}

		b = buf[consumed]
		consumed++
		dist = ((dist + 1) << 7) + uint64(b&0x7f)
	}

	return dist, consumed, nil
}
