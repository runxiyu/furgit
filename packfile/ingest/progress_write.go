package ingest

import "codeberg.org/lindenii/furgit/internal/utils"

func writeProgressf(state *ingestState, format string, args ...any) {
	utils.BestEffortFprintf(state.opts.Progress, format, args...)

	if state.opts.ProgressFlush != nil {
		_ = state.opts.ProgressFlush()
	}
}
