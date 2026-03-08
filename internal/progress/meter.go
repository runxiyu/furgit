package progress

import (
	"io"
	"time"
)

// Meter renders one in-place progress line.
type Meter struct {
	writer io.Writer
	flush  func() error

	title      string
	total      uint64
	delay      time.Duration
	sparse     bool
	throughput bool

	startedAt      time.Time
	nextUpdateAt   time.Time
	nextThroughput time.Time

	lastDone     uint64
	lastBytes    uint64
	lastPercent  int
	lastCounterW int
	sawValue     bool

	throughputSuffix string
}
