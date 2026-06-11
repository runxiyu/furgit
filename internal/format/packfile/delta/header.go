package delta

import (
	"errors"
	"fmt"
)

// ErrMalformedDelta reports that
// a delta payload is truncated, overlong,
// declares sizes that overflow uint64,
// contains invalid instructions,
// or does not match the supplied base or declared sizes.
var ErrMalformedDelta = errors.New("internal/format/packfile/delta: malformed delta")

// MaxSizeVarintLen is the maximum encoded length
// of one delta header size varint.
// Every uint64 size is encodable within this bound,
// and parsing rejects longer encodings.
const MaxSizeVarintLen = 10

// MaxHeaderSizesLen is the maximum encoded length
// of the base/result size header of a delta payload.
//
// Callers reading a delta from a stream may inflate
// MaxHeaderSizesLen bytes
// (or fewer if the payload ends sooner)
// and parse with [ParseHeaderSizes];
// no valid header is longer.
const MaxHeaderSizesLen = 2 * MaxSizeVarintLen

// ParseHeaderSizes parses the base size and result size varints
// at the beginning of one inflated delta payload.
func ParseHeaderSizes(data []byte) (baseSize, resultSize uint64, consumed int, err error) {
	baseSize, consumed, err = parseSizeVarint(data, 0)
	if err != nil {
		return 0, 0, 0, err
	}

	resultSize, consumed, err = parseSizeVarint(data, consumed)
	if err != nil {
		return 0, 0, 0, err
	}

	return baseSize, resultSize, consumed, nil
}

// AppendHeaderSizes appends the base/result size header encoding
// of one delta payload to dst.
func AppendHeaderSizes(dst []byte, baseSize, resultSize uint64) []byte {
	dst = appendSizeVarint(dst, baseSize)

	return appendSizeVarint(dst, resultSize)
}

// appendSizeVarint appends the encoding of one delta header size varint to dst.
func appendSizeVarint(dst []byte, value uint64) []byte {
	for value >= 0x80 {
		dst = append(dst, byte(value)|0x80)
		value >>= 7
	}

	return append(dst, byte(value))
}

// parseSizeVarint parses one delta header size varint
// beginning at pos,
// and returns the position past it.
func parseSizeVarint(data []byte, pos int) (value uint64, consumed int, err error) {
	start := pos

	shift := uint(0)

	for {
		if pos-start >= MaxSizeVarintLen {
			return 0, 0, fmt.Errorf("%w: overlong size varint", ErrMalformedDelta)
		}

		if pos >= len(data) {
			return 0, 0, fmt.Errorf("%w: truncated size varint", ErrMalformedDelta)
		}

		b := data[pos]
		pos++

		group := uint64(b & 0x7f)
		if group<<shift>>shift != group {
			return 0, 0, fmt.Errorf("%w: size overflows uint64", ErrMalformedDelta)
		}

		value |= group << shift

		if b&0x80 == 0 {
			return value, pos, nil
		}

		shift += 7
	}
}
