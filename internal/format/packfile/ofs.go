package packfile

import (
	"errors"
	"fmt"
	"math"
)

// ErrMalformedOfsDeltaDistance reports that
// an ofs-delta backward distance encoding
// is truncated, overlong, or overflows uint64.
var ErrMalformedOfsDeltaDistance = errors.New("internal/format/packfile: malformed ofs-delta distance")

// MaxOfsDeltaDistanceLen is the maximum encoded length
// of an ofs-delta backward distance.
// Every uint64 distance is encodable within this bound,
// and [ParseOfsDeltaDistance] rejects longer encodings.
const MaxOfsDeltaDistanceLen = 10

// ParseOfsDeltaDistance parses an ofs-delta backward distance.
func ParseOfsDeltaDistance(buf []byte) (dist uint64, consumed int, err error) {
	if len(buf) == 0 {
		return 0, 0, fmt.Errorf("%w: truncated", ErrMalformedOfsDeltaDistance)
	}

	b := buf[0]
	dist = uint64(b & 0x7f)

	consumed = 1
	for b&0x80 != 0 {
		if consumed >= MaxOfsDeltaDistanceLen {
			return 0, 0, fmt.Errorf("%w: overlong encoding", ErrMalformedOfsDeltaDistance)
		}

		if consumed >= len(buf) {
			return 0, 0, fmt.Errorf("%w: truncated", ErrMalformedOfsDeltaDistance)
		}

		if dist >= math.MaxUint64>>7 {
			return 0, 0, fmt.Errorf("%w: overflows uint64", ErrMalformedOfsDeltaDistance)
		}

		b = buf[consumed]
		consumed++
		dist = ((dist + 1) << 7) | uint64(b&0x7f)
	}

	return dist, consumed, nil
}

// AppendOfsDeltaDistance appends the encoding of
// an ofs-delta backward distance to dst.
func AppendOfsDeltaDistance(dst []byte, dist uint64) []byte {
	var buf [MaxOfsDeltaDistanceLen]byte

	pos := len(buf) - 1
	buf[pos] = byte(dist & 0x7f)

	dist >>= 7
	for dist != 0 {
		dist--
		pos--
		buf[pos] = 0x80 | byte(dist&0x7f)
		dist >>= 7
	}

	return append(dst, buf[pos:]...)
}
