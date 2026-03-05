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

// IntToUint32 converts v to uint32, returning an error if it overflows.
func IntToUint32(v int) (uint32, error) {
	if v < 0 || v > math.MaxUint32 {
		return 0, fmt.Errorf("intconv: int %d overflows uint32", v)
	}

	return uint32(v), nil
}

// Uint64ToInt64 converts v to int64, returning an error if it overflows.
func Uint64ToInt64(v uint64) (int64, error) {
	if v > math.MaxInt64 {
		return 0, fmt.Errorf("intconv: uint64 %d overflows int64", v)
	}

	return int64(v), nil
}

// Int64ToUint64 converts v to uint64, returning an error if v is negative.
func Int64ToUint64(v int64) (uint64, error) {
	if v < 0 {
		return 0, fmt.Errorf("intconv: int64 %d is negative", v)
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

// SignExtendByteToUint32 sign-extends b as a signed 8-bit integer into uint32.
func SignExtendByteToUint32(b byte) uint32 {
	if b&0x80 == 0 {
		return uint32(b)
	}

	return 0xFFFFFF00 | uint32(b)
}
