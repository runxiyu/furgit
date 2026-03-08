package progress

import "time"

const (
	// DefaultDelay is the default delayed-progress interval.
	DefaultDelay = time.Second

	updateInterval     = time.Second
	throughputInterval = 500 * time.Millisecond
)
