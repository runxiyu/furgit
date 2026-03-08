package progress

import "time"

func (meter *Meter) consumeUpdateTick(now time.Time) bool {
	if now.Before(meter.nextUpdateAt) {
		return false
	}

	for !now.Before(meter.nextUpdateAt) {
		meter.nextUpdateAt = meter.nextUpdateAt.Add(updateInterval)
	}

	return true
}
