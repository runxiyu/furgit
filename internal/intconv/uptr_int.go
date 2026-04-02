package intconv

import (
	"math"
)

// UintptrToInt converts v to int, returning an error if it overflows.
func UintptrToInt(v uintptr) (int, error) {
	if v > uintptr(math.MaxInt) {
		return 0, ErrOverflow
	}

	return int(v), nil
}
