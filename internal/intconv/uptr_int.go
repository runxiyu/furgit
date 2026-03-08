package intconv

import (
	"fmt"
	"math"
)

// UintptrToInt converts v to int, returning an error if it overflows.
func UintptrToInt(v uintptr) (int, error) {
	if v > uintptr(math.MaxInt) {
		return 0, fmt.Errorf("intconv: uintptr %d overflows int", v)
	}

	return int(v), nil
}
