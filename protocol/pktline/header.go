package pktline

import "fmt"

func hexval(b byte) int {
	switch {
	case b >= '0' && b <= '9':
		return int(b - '0')
	case b >= 'a' && b <= 'f':
		return int(b-'a') + 10
	case b >= 'A' && b <= 'F':
		return int(b-'A') + 10
	default:
		return -1
	}
}

// ParseLengthHeader parses a 4-byte hexadecimal pkt-line length header.
//
// The returned value is the full on-wire packet size, including the 4-byte
// header. Semantic interpretation (data/control/error) is done by Decoder.
//
// The 4-byte header is only an actual length when above or equal to 4.
// Otherwise, it indicates some control packet.
func ParseLengthHeader(h [4]byte) (int, error) {
	a := hexval(h[0])
	b := hexval(h[1])
	c := hexval(h[2])
	d := hexval(h[3])

	if a < 0 || b < 0 || c < 0 || d < 0 {
		return 0, fmt.Errorf("%w: %q", ErrInvalidLength, string(h[:]))
	}

	return (a << 12) | (b << 8) | (c << 4) | d, nil
}

// EncodeLengthHeader encodes n as a 4-byte hexadecimal pkt-line header.
//
// n is the full on-wire packet size including the 4-byte header.
//
// The 4-byte header is only an actual length when above or equal to 4.
// Otherwise, it indicates some control packet.
func EncodeLengthHeader(dst *[4]byte, n int) error {
	if n < 0 || n > LargePacketMax {
		return fmt.Errorf("%w: %d", ErrInvalidLength, n)
	}

	const hex = "0123456789abcdef"

	dst[0] = hex[(n>>12)&0xf]
	dst[1] = hex[(n>>8)&0xf]
	dst[2] = hex[(n>>4)&0xf]
	dst[3] = hex[n&0xf]

	return nil
}
