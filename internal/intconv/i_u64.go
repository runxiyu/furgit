package intconv

import "fmt"

// IntToUint64 converts v to uint64, returning an error if v is negative.
func IntToUint64(v int) (uint64, error) {
	if v < 0 {
		return 0, fmt.Errorf("intconv: int %d is negative", v)
	}

	return uint64(v), nil
}
