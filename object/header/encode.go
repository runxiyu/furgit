package header

import "codeberg.org/lindenii/furgit/objecttype"

// Encode returns a canonical loose-object header ("type size\\x00").
func Encode(ty objecttype.Type, size int64) ([]byte, bool) {
	return Append(nil, ty, size)
}
