package ingest

import (
	"fmt"

	"codeberg.org/lindenii/furgit/internal/utils"
)

// ingest initializes transaction state and executes the ingest pipeline.
func ingest(state *ingestState) (out Result, err error) {
	err = openTemporaryArtifacts(state)
	if err != nil {
		return Result{}, err
	}

	defer func() {
		_ = closeTemporaryArtifacts(state)
		if err != nil {
			rollbackTemporaryArtifacts(state)
		}
	}()

	err = streamPackAndScan(state)
	if err != nil {
		return Result{}, err
	}

	err = resolveAll(state)
	if err != nil {
		return Result{}, err
	}

	err = maybeFixThin(state)
	if err != nil {
		return Result{}, err
	}

	if state.thinFixed {
		err := resolveAll(state)
		if err != nil {
			return Result{}, err
		}
	}

	if len(state.unresolvedRefDeltas) > 0 {
		return Result{}, &ThinPackUnresolvedError{Count: len(state.unresolvedRefDeltas)}
	}

	err = verifyResolvedRecords(state)
	if err != nil {
		return Result{}, err
	}

	utils.WriteProgressf(state.opts.Progress, "writing index...\r")

	err = state.packFile.Sync()
	if err != nil {
		return Result{}, &DestinationWriteError{Op: fmt.Sprintf("sync pack: %v", err)}
	}

	err = writeIdx(state)
	if err != nil {
		return Result{}, err
	}

	utils.WriteProgressf(state.opts.Progress, "writing index: done.\n")

	if state.opts.WriteRev {
		utils.WriteProgressf(state.opts.Progress, "writing reverse index...\r")
	}

	err = writeRev(state)
	if err != nil {
		return Result{}, err
	}

	if state.opts.WriteRev {
		utils.WriteProgressf(state.opts.Progress, "writing reverse index: done.\n")
	}

	return finalizeArtifacts(state)
}
