package intconv

import (
	"fmt"
	"math"
)

// Uint32ToUint8 converts v to uint8, returning an error if it overflows.
func Uint32ToUint8(v uint32) (uint8, error) {
	if v > math.MaxUint8 {
		return 0, fmt.Errorf("intconv: uint32 %d overflows uint8", v)
	}

	return uint8(v), nil
}
