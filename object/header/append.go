package header

import (
	"strconv"

	"codeberg.org/lindenii/furgit/object/typ"
)

// AppendHeader appends a canonical loose-object header ("type size\x00") to dst.
func AppendHeader(dst []byte, ty typ.Type, size uint64) ([]byte, bool) {
	tyName, ok := ty.Name()
	if !ok {
		return nil, false
	}

	sizeStr := strconv.FormatInt(size, 10)
	out := make([]byte, 0, len(dst)+len(tyName)+len(sizeStr)+2)
	out = append(out, dst...)
	out = append(out, tyName...)
	out = append(out, ' ')
	out = append(out, sizeStr...)
	out = append(out, 0)

	return out, true
}
