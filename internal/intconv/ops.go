package intconv

import "math"

// Int64ToInt32 converts v to int32,
// returning an error if it overflows.
func Int64ToInt32(v int64) (int32, error) {
	if v < math.MinInt32 || v > math.MaxInt32 {
		return 0, ErrOverflow
	}

	return int32(v), nil
}

// Int64ToUint64 converts v to uint64,
// returning an error if v is negative.
func Int64ToUint64(v int64) (uint64, error) {
	if v < 0 {
		return 0, ErrOverflow
	}

	return uint64(v), nil
}

// IntToUint32 converts v to uint32,
// returning an error if it overflows.
func IntToUint32(v int) (uint32, error) {
	if v < 0 || v > math.MaxUint32 {
		return 0, ErrOverflow
	}

	return uint32(v), nil
}

// IntToUint64 converts v to uint64,
// returning an error if v is negative.
func IntToUint64(v int) (uint64, error) {
	if v < 0 {
		return 0, ErrOverflow
	}

	return uint64(v), nil
}

// SignExtendByteToUint32 sign-extends b
// as a signed 8-bit integer into uint32.
func SignExtendByteToUint32(b byte) uint32 {
	if b&0x80 == 0 {
		return uint32(b)
	}

	return 0xFFFFFF00 | uint32(b)
}

// Uint32ToInt converts v to int,
// returning an error if it overflows.
func Uint32ToInt(v uint32) (int, error) {
	if uint64(v) > uint64(math.MaxInt) {
		return 0, ErrOverflow
	}

	return int(v), nil
}

// Uint32ToUint8 converts v to uint8,
// returning an error if it overflows.
func Uint32ToUint8(v uint32) (uint8, error) {
	if v > math.MaxUint8 {
		return 0, ErrOverflow
	}

	return uint8(v), nil
}

// Uint64ToInt converts v to int,
// returning an error if it overflows.
func Uint64ToInt(v uint64) (int, error) {
	if v > uint64(math.MaxInt) {
		return 0, ErrOverflow
	}

	return int(v), nil
}

// Uint64ToInt64 converts v to int64,
// returning an error if it overflows.
func Uint64ToInt64(v uint64) (int64, error) {
	if v > math.MaxInt64 {
		return 0, ErrOverflow
	}

	return int64(v), nil
}

// UintptrToInt converts v to int,
// returning an error if it overflows.
func UintptrToInt(v uintptr) (int, error) {
	if v > uintptr(math.MaxInt) {
		return 0, ErrOverflow
	}

	return int(v), nil
}
