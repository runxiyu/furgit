package header

import (
	"slices"
	"strconv"

	"lindenii.org/go/furgit/object/typ"
)

// Append appends a canonical loose-object header ("type size\x00") to dst.
func Append(dst []byte, ty typ.Type, size uint64) []byte {
	tyName := ty.Name()

	dst = slices.Grow(dst, len(tyName)+1+19+1)
	dst = append(dst, tyName...)
	dst = append(dst, ' ')
	dst = strconv.AppendUint(dst, size, 10)
	dst = append(dst, 0)

	return dst
}
