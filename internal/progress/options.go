package progress

import (
	"time"

	"lindenii.org/go/furgit/common/iowrap"
)

// Options configures one progress meter.
type Options struct {
	Writer iowrap.WriteFlusher

	Title string
	Total uint64

	// Delay suppresses progress output until Delay has elapsed since Start.
	Delay time.Duration
	// Sparse forces one final 100% line at Stop when the caller sampled updates.
	Sparse bool
	// Throughput appends ", <total> | <rate>/s" and refreshes rate every 500ms.
	Throughput bool
}
