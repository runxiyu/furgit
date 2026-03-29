package objectid

// HexLen returns the encoded hexadecimal length.
func (algo Algorithm) HexLen() int {
	return algo.Size() * 2
}
