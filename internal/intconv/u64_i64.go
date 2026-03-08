package intconv

import (
	"fmt"
	"math"
)

// Uint64ToInt64 converts v to int64, returning an error if it overflows.
func Uint64ToInt64(v uint64) (int64, error) {
	if v > math.MaxInt64 {
		return 0, fmt.Errorf("intconv: uint64 %d overflows int64", v)
	}

	return int64(v), nil
}
