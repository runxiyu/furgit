package progress

import "time"

// New creates one progress meter.
func New(opts Options) *Meter {
	now := time.Now()

	return &Meter{
		writer:         opts.Writer,
		flush:          opts.Flush,
		title:          opts.Title,
		total:          opts.Total,
		delay:          max(opts.Delay, time.Duration(0)),
		sparse:         opts.Sparse,
		throughput:     opts.Throughput,
		startedAt:      now,
		nextUpdateAt:   now.Add(updateInterval),
		nextThroughput: now.Add(throughputInterval),
		lastPercent:    -1,
	}
}
