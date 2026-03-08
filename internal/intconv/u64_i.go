package intconv

import (
	"fmt"
	"math"
)

// Uint64ToInt converts v to int, returning an error if it overflows.
func Uint64ToInt(v uint64) (int, error) {
	if v > uint64(math.MaxInt) {
		return 0, fmt.Errorf("intconv: uint64 %d overflows int", v)
	}

	return int(v), nil
}
