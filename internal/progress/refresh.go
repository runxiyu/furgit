package progress

import "time"

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
	meter.throughputSuffix = ", " + humanizeBytes(meter.lastBytes) + " | " + humanizeBytes(rate) + "/s"
}
