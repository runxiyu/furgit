package intconv

import (
	"math"
)

// Uint64ToInt converts v to int, returning an error if it overflows.
func Uint64ToInt(v uint64) (int, error) {
	if v > uint64(math.MaxInt) {
		return 0, ErrOverflow
	}

	return int(v), nil
}
