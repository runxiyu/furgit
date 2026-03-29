package objectid

// PackHashID returns the Git pack/rev hash-id encoding for this algorithm.
//
// Unknown algorithms return 0.
func (algo Algorithm) PackHashID() uint32 {
	return algo.info().packHashID
}
