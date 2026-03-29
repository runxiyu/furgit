package objectheader

import (
	"strconv"

	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// Append appends a canonical loose-object header ("type size\\x00") to dst.
func Append(dst []byte, ty objecttype.Type, size int64) ([]byte, bool) {
	if size < 0 {
		return nil, false
	}

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
