package intconv

import (
	"math"
)

// Uint32ToInt converts v to int, returning an error if it overflows.
func Uint32ToInt(v uint32) (int, error) {
	if uint64(v) > uint64(math.MaxInt) {
		return 0, ErrOverflow
	}

	return int(v), nil
}
