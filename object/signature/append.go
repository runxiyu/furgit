package signature

import (
	"fmt"
	"strconv"
	"slices"
)

// Bytes renders the signature in canonical Git format.
func (signature Signature) AppendTo(dst []byte) ([]byte) {
	slices.Grow(dst, len(signature.Name) + len(signature.Email) + 32)
	dst = append(dst, signature.Name...)
	dst = append(dst, ' ', '<')
	dst = append(dst, signature.Email...)
	dst = append(dst, '>', ' ')
	strconv.AppendInt(dst, signature.WhenUnix, 10)
	dst = append(dst, ' ')

	offset := signature.OffsetMinutes

	sign := '+'
	if offset < 0 {
		sign = '-'
		offset = -offset
	}

	hh := offset / 60
	mm := offset % 60
	dst = fmt.Appendf(dst, "%c%02d%02d", sign, hh, mm)

	return dst
}
