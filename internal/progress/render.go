package progress

import (
	"strings"
	"time"

	"lindenii.org/go/furgit/internal/utils"
)

func (meter *Meter) render(now time.Time, eol string) {
	if meter.delay > 0 && now.Sub(meter.startedAt) < meter.delay && eol == "\r" {
		return
	}

	meter.refreshThroughput(now)

	counters := meter.renderCounters()

	clear1 := 0
	if len(counters) < meter.lastCounterW {
		clear1 = meter.lastCounterW - len(counters) + 1
	}

	meter.lastCounterW = len(counters)

	line := meter.title + ": " + counters
	if clear1 > 0 {
		line += strings.Repeat(" ", clear1)
	}

	line += eol

	utils.BestEffortFprintf(meter.writer, "%s", line)

	if meter.writer != nil {
		_ = meter.writer.Flush()
	}
}
