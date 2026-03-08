package intconv

import "fmt"

// Int64ToUint64 converts v to uint64, returning an error if v is negative.
func Int64ToUint64(v int64) (uint64, error) {
	if v < 0 {
		return 0, fmt.Errorf("intconv: int64 %d is negative", v)
	}

	return uint64(v), nil
}
