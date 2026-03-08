package progress

import (
	"strings"
	"time"

	"codeberg.org/lindenii/furgit/internal/utils"
)

func (meter *Meter) render(now time.Time, eol string) {
	if meter.delay > 0 && now.Sub(meter.startedAt) < meter.delay && eol == "\r" {
		return
	}

	meter.refreshThroughput(now)

	counters := meter.renderCounters()
	clear := 0
	if len(counters) < meter.lastCounterW {
		clear = meter.lastCounterW - len(counters) + 1
	}
	meter.lastCounterW = len(counters)

	line := meter.title + ": " + counters
	if clear > 0 {
		line += strings.Repeat(" ", clear)
	}
	line += eol

	utils.BestEffortFprintf(meter.writer, "%s", line)
	if meter.flush != nil {
		_ = meter.flush()
	}
}
