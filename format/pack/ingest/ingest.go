package ingest

import "fmt"

// ingest initializes transaction state and executes the ingest pipeline.
func ingest(state *ingestState) (out Result, err error) {
	if err := openTemporaryArtifacts(state); err != nil {
		return Result{}, err
	}

	defer func() {
		_ = closeTemporaryArtifacts(state)
		if err != nil {
			rollbackTemporaryArtifacts(state)
		}
	}()

	if err := streamPackAndScan(state); err != nil {
		return Result{}, err
	}

	if err := resolveAll(state); err != nil {
		return Result{}, err
	}

	if err := maybeFixThin(state); err != nil {
		return Result{}, err
	}

	if state.thinFixed {
		if err := resolveAll(state); err != nil {
			return Result{}, err
		}
	}

	if len(state.unresolvedRefDeltas) > 0 {
		return Result{}, &ErrThinPackUnresolved{Count: len(state.unresolvedRefDeltas)}
	}

	if err := verifyResolvedRecords(state); err != nil {
		return Result{}, err
	}

	if err := state.packFile.Sync(); err != nil {
		return Result{}, &ErrDestinationWrite{Op: fmt.Sprintf("sync pack: %v", err)}
	}

	if err := writeIdx(state); err != nil {
		return Result{}, err
	}

	if err := writeRev(state); err != nil {
		return Result{}, err
	}

	return finalizeArtifacts(state)
}
