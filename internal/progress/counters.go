package progress

import (
	"fmt"

	"codeberg.org/lindenii/furgit/internal/intconv"
)

func (meter *Meter) renderCounters() string {
	if meter.total > 0 {
		u, err := intconv.Uint64ToInt(meter.lastDone * 100 / meter.total)
		if err != nil {
			return "overflow"
			// TODO
		}

		meter.lastPercent = u

		return fmt.Sprintf("%3d%% (%d/%d)%s", meter.lastPercent, meter.lastDone, meter.total, meter.throughputSuffix)
	}

	return fmt.Sprintf("%d%s", meter.lastDone, meter.throughputSuffix)
}
