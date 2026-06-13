package progress

import (
	"time"

	"lindenii.org/go/lgo/iowrap"
)

const (
	updateInterval     = time.Second
	throughputInterval = 500 * time.Millisecond
)

// Meter renders one in-place progress line.
type Meter struct {
	writer iowrap.WriteFlusher

	title      string
	total      int
	delay      time.Duration
	sparse     bool
	throughput bool

	startedAt      time.Time
	nextUpdateAt   time.Time
	nextThroughput time.Time

	lastDone     int
	lastBytes    int
	lastPercent  int
	lastCounterW int
	sawValue     bool

	throughputSuffix string
}

// New creates one progress meter.
func New(opts Options) *Meter {
	now := time.Now()

	return &Meter{
		writer:         opts.Writer,
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

// Options configures one progress meter.
type Options struct {
	Writer iowrap.WriteFlusher

	Title string
	Total int

	// Delay suppresses progress output until Delay has elapsed since Start.
	Delay time.Duration
	// Sparse forces one final 100% line at Stop when the caller sampled updates.
	Sparse bool
	// Throughput appends ", <total> | <rate>/s" and refreshes rate every 500ms.
	Throughput bool
}

// Set records current progress
// and renders when percent changed or the 1s tick elapsed.
func (meter *Meter) Set(done int, bytes int) {
	meter.lastDone = done
	meter.lastBytes = bytes
	meter.sawValue = true

	if meter.writer == nil {
		return
	}

	now := time.Now()
	forced := meter.consumeUpdateTick(now)

	percentChanged := false

	if meter.total > 0 {
		percent := int(int64(done) * 100 / int64(meter.total))
		percentChanged = percent != meter.lastPercent
	}

	if !percentChanged && !forced {
		return
	}

	meter.render(now, "\r")
}

// Stop forces the final progress line and appends ", <msg>.".
func (meter *Meter) Stop(msg string) {
	if !meter.sawValue || meter.writer == nil {
		return
	}

	if msg == "" {
		msg = "done"
	}

	if meter.sparse && meter.total > 0 && meter.lastDone != meter.total {
		meter.lastDone = meter.total
	}

	meter.render(time.Now(), ", "+msg+".\n")
}

func (meter *Meter) consumeUpdateTick(now time.Time) bool {
	if now.Before(meter.nextUpdateAt) {
		return false
	}

	for !now.Before(meter.nextUpdateAt) {
		meter.nextUpdateAt = meter.nextUpdateAt.Add(updateInterval)
	}

	return true
}
