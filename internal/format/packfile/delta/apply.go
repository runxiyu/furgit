package delta

import (
	"fmt"

	"lindenii.org/go/lgo/intconv"
)

// MaxChainDepth is the maximum supported delta chain length.
// Resolvers reject chains deeper than this
// to bound recursion depth and reconstruction work.
const MaxChainDepth = 1 << 12

// Apply applies one inflated delta payload to base
// and returns the reconstructed result.
//
// delta must include the leading base/result size header.
func Apply(base, delta []byte) ([]byte, error) {
	baseSize, resultSize, pos, err := ParseHeaderSizes(delta)
	if err != nil {
		return nil, err
	}

	if baseSize != uint64(len(base)) {
		return nil, fmt.Errorf("%w: base size mismatch", ErrMalformedDelta)
	}

	outLen, err := intconv.Uint64ToInt(resultSize)
	if err != nil {
		return nil, fmt.Errorf("%w: result size overflows int", ErrMalformedDelta)
	}

	out := make([]byte, outLen)
	outPos := 0

	for pos < len(delta) {
		op := delta[pos]
		pos++

		switch {
		case op&0x80 != 0:
			outPos, err = applyCopy(out, outPos, base, delta, &pos, op)
		case op != 0:
			outPos, err = applyInsert(out, outPos, delta, &pos, int(op))
		default:
			err = fmt.Errorf("%w: invalid opcode 0", ErrMalformedDelta)
		}

		if err != nil {
			return nil, err
		}
	}

	if outPos != len(out) {
		return nil, fmt.Errorf("%w: result size mismatch", ErrMalformedDelta)
	}

	return out, nil
}

// applyCopy executes one copy instruction,
// copying a base range into out,
// and returns the new output position.
func applyCopy(out []byte, outPos int, base, delta []byte, pos *int, op byte) (int, error) {
	off, err := parseCopyOperand(delta, pos, op, 0, 4)
	if err != nil {
		return 0, err
	}

	n, err := parseCopyOperand(delta, pos, op, 4, 3)
	if err != nil {
		return 0, err
	}

	if n == 0 {
		n = 0x10000
	}

	if off+n > len(base) || outPos+n > len(out) {
		return 0, fmt.Errorf("%w: copy out of bounds", ErrMalformedDelta)
	}

	copy(out[outPos:outPos+n], base[off:off+n])

	return outPos + n, nil
}

// applyInsert executes one insert instruction,
// copying n literal delta bytes into out,
// and returns the new output position.
func applyInsert(out []byte, outPos int, delta []byte, pos *int, n int) (int, error) {
	if *pos+n > len(delta) || outPos+n > len(out) {
		return 0, fmt.Errorf("%w: insert out of bounds", ErrMalformedDelta)
	}

	copy(out[outPos:outPos+n], delta[*pos:*pos+n])
	*pos += n

	return outPos + n, nil
}

// parseCopyOperand assembles one little-endian copy instruction operand
// from the operand bytes selected by op's flag bits
// firstBit through firstBit+count-1.
func parseCopyOperand(delta []byte, pos *int, op byte, firstBit uint, count int) (int, error) {
	value := 0

	for i := range count {
		if op&(1<<(firstBit+uint(i))) == 0 { //#nosec G115
			continue
		}

		if *pos >= len(delta) {
			return 0, fmt.Errorf("%w: truncated copy operand", ErrMalformedDelta)
		}

		value |= int(delta[*pos]) << (8 * i)
		*pos++
	}

	return value, nil
}
