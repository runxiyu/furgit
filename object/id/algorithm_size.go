package objectid

// Size returns the hash size in bytes.
func (algo Algorithm) Size() int {
	return algo.info().size
}
