package signature

import (
	"fmt"
	"strconv"
	"strings"
)

// Bytes renders the signature in canonical Git format.
func (signature Signature) Bytes() ([]byte, error) {
	var b strings.Builder
	b.Grow(len(signature.Name) + len(signature.Email) + 32)
	b.Write(signature.Name)
	b.WriteString(" <")
	b.Write(signature.Email)
	b.WriteString("> ")
	b.WriteString(strconv.FormatInt(signature.WhenUnix, 10))
	b.WriteByte(' ')

	offset := signature.OffsetMinutes

	sign := '+'
	if offset < 0 {
		sign = '-'
		offset = -offset
	}

	hh := offset / 60
	mm := offset % 60
	fmt.Fprintf(&b, "%c%02d%02d", sign, hh, mm)

	return []byte(b.String()), nil
}
