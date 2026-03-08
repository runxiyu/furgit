package progress

import "time"

// Stop forces the final progress line and appends ", <msg>.".
func (meter *Meter) Stop(msg string) {
	if !meter.sawValue || meter.writer == nil {
		return
	}

	if msg == "" {
		msg = "done"
	}

	if meter.sparse && meter.total > 0 && meter.lastDone != meter.total {
		meter.lastDone = meter.total
	}

	meter.render(time.Now(), ", "+msg+".\n")
}
