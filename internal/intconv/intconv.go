// Package intconv provides checked integer conversion helpers.
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

// UintptrToInt converts v to int, returning an error if it overflows.
func UintptrToInt(v uintptr) (int, error) {
	if v > uintptr(math.MaxInt) {
		return 0, fmt.Errorf("intconv: uintptr %d overflows int", v)
	}

	return int(v), nil
}

// IntToUint64 converts v to uint64, returning an error if v is negative.
func IntToUint64(v int) (uint64, error) {
	if v < 0 {
		return 0, fmt.Errorf("intconv: int %d is negative", v)
	}

	return uint64(v), nil
}

// Int64ToInt32 converts v to int32, returning an error if it overflows.
func Int64ToInt32(v int64) (int32, error) {
	if v < math.MinInt32 || v > math.MaxInt32 {
		return 0, fmt.Errorf("intconv: int64 %d overflows int32", v)
	}

	return int32(v), nil
}
