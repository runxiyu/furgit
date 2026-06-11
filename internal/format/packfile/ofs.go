package packfile

import (
	"errors"
)

var ErrMalformedOfsDeltaDistance = errors.New("internal/format/packfile/delta: malformed ofs-delta distance")

// ParseOfsDeltaDistance parses an ofs-delta backward distance.
func ParseOfsDeltaDistance(buf []byte) (dist uint64, consumed int, err error) {
	if len(buf) == 0 {
		return 0, 0, ErrMalformedOfsDeltaDistance
	}

	b := buf[0]
	dist = uint64(b & 0x7f)

	consumed = 1
	for b&0x80 != 0 {
		if consumed >= len(buf) {
			return 0, 0, ErrMalformedOfsDeltaDistance
		}

		b = buf[consumed]
		consumed++
		dist = ((dist + 1) << 7) + uint64(b&0x7f)
	}

	return dist, consumed, nil
}
