package progress

import "time"

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
		percent := int(done * 100 / meter.total)
		percentChanged = percent != meter.lastPercent
	}

	if !percentChanged && !forced {
		return
	}

	meter.render(now, "\r")
}
