package progress

import "fmt"

func (meter *Meter) renderCounters() string {
	if meter.total > 0 {
		percent := int(meter.lastDone * 100 / meter.total)
		meter.lastPercent = percent

		return fmt.Sprintf("%3d%% (%d/%d)%s", percent, meter.lastDone, meter.total, meter.throughputSuffix)
	}

	return fmt.Sprintf("%d%s", meter.lastDone, meter.throughputSuffix)
}
