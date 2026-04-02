package intconv

// IntToUint64 converts v to uint64, returning an error if v is negative.
func IntToUint64(v int) (uint64, error) {
	if v < 0 {
		return 0, ErrOverflow
	}

	return uint64(v), nil
}
