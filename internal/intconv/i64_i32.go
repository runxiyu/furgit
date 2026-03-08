package intconv

import (
	"fmt"
	"math"
)

// Int64ToInt32 converts v to int32, returning an error if it overflows.
func Int64ToInt32(v int64) (int32, error) {
	if v < math.MinInt32 || v > math.MaxInt32 {
		return 0, fmt.Errorf("intconv: int64 %d overflows int32", v)
	}

	return int32(v), nil
}
