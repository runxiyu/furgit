package signature

import (
	"slices"
	"strconv"

	"lindenii.org/go/lgo/intconv"
)

// Append renders the signature in canonical Git format.
func (signature Signature) Append(dst []byte) ([]byte, error) {
	dst = slices.Grow(dst, len(signature.Name)+len(signature.Email)+32)
	dst = append(dst, signature.Name...)
	dst = append(dst, ' ', '<')
	dst = append(dst, signature.Email...)
	dst = append(dst, '>', ' ')
	dst = strconv.AppendInt(dst, signature.WhenUnix, 10)
	dst = append(dst, ' ')

	offset := signature.OffsetMinutes
	if offset < -(23*60+59) || offset > 23*60+59 {
		return dst, ErrInvalidSignature
	}

	var sign byte = '+'
	if offset < 0 {
		sign = '-'
		offset = -offset
	}

	hh := offset / 60
	mm := offset % 60

	hhTens, err := intconv.Int32ToUint8('0' + hh/10)
	if err != nil {
		return dst, ErrInvalidSignature
	}

	hhOnes, err := intconv.Int32ToUint8('0' + hh%10)
	if err != nil {
		return dst, ErrInvalidSignature
	}

	mmTens, err := intconv.Int32ToUint8('0' + mm/10)
	if err != nil {
		return dst, ErrInvalidSignature
	}

	mmOnes, err := intconv.Int32ToUint8('0' + mm%10)
	if err != nil {
		return dst, ErrInvalidSignature
	}

	dst = append(dst, sign)
	dst = append(dst, hhTens, hhOnes)
	dst = append(dst, mmTens, mmOnes)

	return dst, nil
}
