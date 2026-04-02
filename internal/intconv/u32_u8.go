package intconv

import (
	"math"
)

// Uint32ToUint8 converts v to uint8, returning an error if it overflows.
func Uint32ToUint8(v uint32) (uint8, error) {
	if v > math.MaxUint8 {
		return 0, ErrOverflow
	}

	return uint8(v), nil
}
