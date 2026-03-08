package intconv

import (
	"fmt"
	"math"
)

// Uint32ToInt converts v to int, returning an error if it overflows.
func Uint32ToInt(v uint32) (int, error) {
	if uint64(v) > uint64(math.MaxInt) {
		return 0, fmt.Errorf("intconv: uint32 %d overflows int", v)
	}

	return int(v), nil
}
