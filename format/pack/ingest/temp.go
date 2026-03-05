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

	idxName, idxFile, err := createTempFile(state.destination, "tmp_idx_")
	if err != nil {
		_ = packFile.Close()
		_ = state.destination.Remove(packName)

		return err
	}

	revName := ""

	var revFile *os.File
	if state.writeRev {
		revName, revFile, err = createTempFile(state.destination, "tmp_rev_")
		if err != nil {
			_ = idxFile.Close()
			_ = state.destination.Remove(idxName)
			_ = packFile.Close()
			_ = state.destination.Remove(packName)

			return err
		}
	}

	state.packTmpName = packName
	state.packFile = packFile
	state.idxTmpName = idxName
	state.idxFile = idxFile
	state.revTmpName = revName
	state.revFile = revFile

	return nil
}

// closeTemporaryArtifacts closes all temporary artifact file descriptors.
func closeTemporaryArtifacts(state *ingestState) error {
	var out error

	if state.packFile != nil {
		err := state.packFile.Close()
		if err != nil && out == nil {
			out = err
		}

		state.packFile = nil
	}

	if state.idxFile != nil {
		err := state.idxFile.Close()
		if err != nil && out == nil {
			out = err
		}

		state.idxFile = nil
	}

	if state.revFile != nil {
		err := state.revFile.Close()
		if err != nil && out == nil {
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
