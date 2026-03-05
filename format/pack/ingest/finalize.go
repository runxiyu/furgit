package ingest

import (
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"codeberg.org/lindenii/furgit/internal/intconv"
)

// finalizeArtifacts links temporary files to final names and returns Result.
func finalizeArtifacts(state *ingestState) (Result, error) {
	base := "pack-" + state.packHash.String()
	packFinal := base + ".pack"
	idxFinal := base + ".idx"

	revFinal := ""
	if state.writeRev {
		revFinal = base + ".rev"
	}

	err := linkTempToFinal(state, state.packTmpName, packFinal)
	if err != nil {
		return Result{}, err
	}

	err = linkTempToFinal(state, state.idxTmpName, idxFinal)
	if err != nil {
		return Result{}, err
	}

	if state.writeRev {
		err := linkTempToFinal(state, state.revTmpName, revFinal)
		if err != nil {
			return Result{}, err
		}
	}

	objectCount, err := intconv.IntToUint32(len(state.records))
	if err != nil {
		return Result{}, err
	}

	return Result{
		PackName:    packFinal,
		IdxName:     idxFinal,
		RevName:     revFinal,
		PackHash:    state.packHash,
		ObjectCount: objectCount,
		ThinFixed:   state.thinFixed,
	}, nil
}

// rollbackTemporaryArtifacts removes temporary files after failure.
func rollbackTemporaryArtifacts(state *ingestState) {
	if state.packTmpName != "" {
		_ = state.destination.Remove(state.packTmpName)
	}

	if state.idxTmpName != "" {
		_ = state.destination.Remove(state.idxTmpName)
	}

	if state.revTmpName != "" {
		_ = state.destination.Remove(state.revTmpName)
	}
}

// linkTempToFinal hard-links tmp to final, tolerating existing final paths.
func linkTempToFinal(state *ingestState, tmp, final string) error {
	if tmp == "" || final == "" {
		return fmt.Errorf("format/pack/ingest: invalid finalize names tmp=%q final=%q", tmp, final)
	}

	if strings.Contains(final, "/") {
		return fmt.Errorf("format/pack/ingest: final name must be leaf: %q", final)
	}

	err := state.destination.Link(tmp, final)
	if err == nil {
		_ = state.destination.Remove(tmp)

		return nil
	}

	if errors.Is(err, fs.ErrExist) {
		_ = state.destination.Remove(tmp)

		return nil
	}

	return err
}
