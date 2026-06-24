package progress

import (
	"sync/atomic"
	"time"

	"lindenii.org/go/lgo/iowrap"
)

const (
	renderInterval     = 100 * time.Millisecond
	forceInterval      = time.Second
	throughputInterval = 500 * time.Millisecond
)

// Meter renders one in-place progress line.
//
// Add is safe for concurrent use; a single background goroutine renders.
// Stop must be called exactly once to flush the final line and release it.
type Meter struct {
	writer iowrap.WriteFlusher

	title      string
	total      int
	delay      time.Duration
	sparse     bool
	throughput bool

	done     atomic.Int64
	bytes    atomic.Int64
	sawValue atomic.Bool

	startedAt time.Time

	stop   chan struct{}
	exited chan struct{}

	// The following are owned by the render goroutine while it runs,
	// then by Stop once exited is closed.
	nextForceAt      time.Time
	nextThroughput   time.Time
	lastPercent      int
	lastCounterW     int
	throughputSuffix string
}

// New creates one progress meter and starts its render goroutine.
func New(opts Options) *Meter {
	now := time.Now()

	meter := &Meter{
		writer:         opts.Writer,
		title:          opts.Title,
		total:          opts.Total,
		delay:          max(opts.Delay, time.Duration(0)),
		sparse:         opts.Sparse,
		throughput:     opts.Throughput,
		startedAt:      now,
		stop:           make(chan struct{}),
		exited:         make(chan struct{}),
		nextForceAt:    now.Add(forceInterval),
		nextThroughput: now.Add(throughputInterval),
		lastPercent:    -1,
	}

	if meter.writer != nil {
		go meter.loop()
	} else {
		close(meter.exited)
	}

	return meter
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

// Add increments the done and byte counters.
//
// Labels: MT-Safe.
func (meter *Meter) Add(done, bytes int64) {
	meter.done.Add(done)
	meter.bytes.Add(bytes)
	meter.sawValue.Store(true)
}

// Stop ends the render goroutine, forces the final line, and appends ", <msg>.".
func (meter *Meter) Stop(msg string) {
	close(meter.stop)
	<-meter.exited

	if !meter.sawValue.Load() || meter.writer == nil {
		return
	}

	if msg == "" {
		msg = "done"
	}

	if meter.sparse && meter.total > 0 && int(meter.done.Load()) != meter.total {
		meter.done.Store(int64(meter.total))
	}

	meter.render(time.Now(), ", "+msg+".\n")
}

func (meter *Meter) loop() {
	defer close(meter.exited)

	ticker := time.NewTicker(renderInterval)
	defer ticker.Stop()

	for {
		select {
		case <-meter.stop:
			return
		case now := <-ticker.C:
			meter.maybeRender(now)
		}
	}
}

func (meter *Meter) maybeRender(now time.Time) {
	if !meter.sawValue.Load() {
		return
	}

	forced := false

	for !now.Before(meter.nextForceAt) {
		meter.nextForceAt = meter.nextForceAt.Add(forceInterval)
		forced = true
	}

	percentChanged := false

	if meter.total > 0 {
		percent := int(meter.done.Load() * 100 / int64(meter.total))
		percentChanged = percent != meter.lastPercent
	}

	if percentChanged || forced {
		meter.render(now, "\r")
	}
}
