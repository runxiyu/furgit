package ingest

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io/fs"
	"os"
)

// openTemporaryArtifacts creates/open temp pack/idx/(rev) files under destination.
func openTemporaryArtifacts(state *ingestState) error {
	packName, packFile, err := createTempFile(state.destination, "tmp_pack_")
	if err != nil {
		return err
	}
	state.packTmpName = packName
	state.packFile = packFile

	idxName, idxFile, err := createTempFile(state.destination, "tmp_idx_")
	if err != nil {
		return err
	}
	state.idxTmpName = idxName
	state.idxFile = idxFile

	if state.writeRev {
		revName, revFile, err := createTempFile(state.destination, "tmp_rev_")
		if err != nil {
			return err
		}
		state.revTmpName = revName
		state.revFile = revFile
	}

	return nil
}

// closeTemporaryArtifacts closes all temporary artifact file descriptors.
func closeTemporaryArtifacts(state *ingestState) error {
	var out error
	if state.packFile != nil {
		if err := state.packFile.Close(); err != nil && out == nil {
			out = err
		}
		state.packFile = nil
	}
	if state.idxFile != nil {
		if err := state.idxFile.Close(); err != nil && out == nil {
			out = err
		}
		state.idxFile = nil
	}
	if state.revFile != nil {
		if err := state.revFile.Close(); err != nil && out == nil {
			out = err
		}
		state.revFile = nil
	}

	return out
}

// createTempFile creates one temporary file under root using prefix.
func createTempFile(root *os.Root, prefix string) (string, *os.File, error) {
	for range 32 {
		name := prefix + rand.Text()
		file, err := root.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0o644)
		if err == nil {
			return name, file, nil
		}
		if errors.Is(err, fs.ErrExist) {
			continue
		}

		return "", nil, fmt.Errorf("format/pack/ingest: create temp file %q: %w", name, err)
	}

	return "", nil, fmt.Errorf("format/pack/ingest: unable to create temporary file for prefix %q", prefix)
}
