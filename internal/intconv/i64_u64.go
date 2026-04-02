package intconv

// Int64ToUint64 converts v to uint64, returning an error if v is negative.
func Int64ToUint64(v int64) (uint64, error) {
	if v < 0 {
		return 0, ErrOverflow
	}

	return uint64(v), nil
}
