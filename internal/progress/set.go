package progress

import (
	"time"

	"codeberg.org/lindenii/furgit/internal/intconv"
)

// Set records current progress and renders when percent changed or the 1s tick
// elapsed.
func (meter *Meter) Set(done uint64, bytes uint64) {
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
		percent, err := intconv.Uint64ToInt(done * 100 / meter.total)
		if err != nil {
			return // TODO
		}

		percentChanged = percent != meter.lastPercent
	}

	if !percentChanged && !forced {
		return
	}

	meter.render(now, "\r")
}
