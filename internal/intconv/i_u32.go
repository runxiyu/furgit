package intconv

import (
	"math"
)

// IntToUint32 converts v to uint32, returning an error if it overflows.
func IntToUint32(v int) (uint32, error) {
	if v < 0 || v > math.MaxUint32 {
		return 0, ErrOverflow
	}

	return uint32(v), nil
}
