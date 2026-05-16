package signature

import (
	"slices"
	"strconv"
)

// AppendTo renders the signature in canonical Git format.
func (signature Signature) AppendTo(dst []byte) []byte {
	dst = slices.Grow(dst, len(signature.Name)+len(signature.Email)+32)
	dst = append(dst, signature.Name...)
	dst = append(dst, ' ', '<')
	dst = append(dst, signature.Email...)
	dst = append(dst, '>', ' ')
	dst = strconv.AppendInt(dst, signature.WhenUnix, 10)
	dst = append(dst, ' ')

	offset := signature.OffsetMinutes

	sign := '+'
	if offset < 0 {
		sign = '-'
		offset = -offset
	}

	hh := offset / 60
	mm := offset % 60

	dst = append(dst, byte(sign))
	dst = append(dst, byte('0'+hh/10), byte('0'+hh%10))
	dst = append(dst, byte('0'+mm/10), byte('0'+mm%10))

	return dst
}
