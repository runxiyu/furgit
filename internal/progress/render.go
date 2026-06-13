package progress

import (
	"fmt"
	"strings"
	"time"

	"lindenii.org/go/furgit/internal/utils"
	"lindenii.org/go/lgo/fmt/humanize"
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

func (meter *Meter) renderCounters() string {
	if meter.total > 0 {
		meter.lastPercent = int(int64(meter.lastDone) * 100 / int64(meter.total))

		return fmt.Sprintf("%3d%% (%d/%d)%s", meter.lastPercent, meter.lastDone, meter.total, meter.throughputSuffix)
	}

	return fmt.Sprintf("%d%s", meter.lastDone, meter.throughputSuffix)
}

func (meter *Meter) refreshThroughput(now time.Time) {
	if !meter.throughput {
		return
	}

	if meter.nextThroughput.After(now) && meter.throughputSuffix != "" {
		return
	}

	for !now.Before(meter.nextThroughput) {
		meter.nextThroughput = meter.nextThroughput.Add(throughputInterval)
	}

	elapsed := now.Sub(meter.startedAt)
	if elapsed <= 0 {
		return
	}

	rate := uint64(float64(meter.lastBytes) / elapsed.Seconds())
	meter.throughputSuffix = ", " + humanize.Bytes(uint64(meter.lastBytes)) + " | " + humanize.Bytes(rate) + "/s" //nolint:gosec
}
