package objectid

import "bytes"

// Compare lexicographically compares two object IDs by their canonical byte
// representation.
func Compare(left, right ObjectID) int {
	return bytes.Compare(left.RawBytes(), right.RawBytes())
}
