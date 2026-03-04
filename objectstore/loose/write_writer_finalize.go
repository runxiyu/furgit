package loose

import (
	"errors"
	"io/fs"
	"path/filepath"

	"codeberg.org/lindenii/furgit/objectid"
)

// Close flushes and closes the underlying zlib stream and temp file.
// It is safe to call multiple times.
func (writer *streamWriter) Close() error {
	if writer.closed {
		return nil
	}

	writer.closed = true

	errZlib := writer.zw.Close()
	errSync := writer.file.Sync()
	errFile := writer.file.Close()
	writer.file = nil

	return errors.Join(errZlib, errSync, errFile)
}

// finalize validates write completeness and atomically publishes the object.
// Publication is no-clobber: it links tmpRelPath to the object path and treats
// existing destination objects as success.
func (writer *streamWriter) finalize() (objectid.ObjectID, error) {
	if writer.finalized {
		return writer.finalID, writer.finalErr
	}

	writer.finalized = true

	var zero objectid.ObjectID

	if !writer.closed {
		err := writer.Close()
		if err != nil {
			writer.finalErr = err

			return zero, err
		}
	}

	if writer.fullMode && !writer.headerDone {
		writer.finalErr = errors.New("objectstore/loose: missing full object header")

		return zero, writer.finalErr
	}

	if writer.expectedContentLeft != 0 {
		writer.finalErr = errors.New("objectstore/loose: object content shorter than declared size")

		return zero, writer.finalErr
	}

	idBytes := writer.hash.Sum(nil)

	id, err := objectid.FromBytes(writer.store.algo, idBytes)
	if err != nil {
		writer.finalErr = err

		return zero, err
	}

	relPath, err := writer.store.objectPath(id)
	if err != nil {
		writer.finalErr = err

		return zero, err
	}

	dir := filepath.Dir(relPath)

	err = writer.store.root.MkdirAll(dir, 0o755)
	if err != nil {
		writer.finalErr = err

		return zero, err
	}

	cleanup := true

	defer func() {
		if cleanup {
			_ = writer.store.root.Remove(writer.tmpRelPath)
		}
	}()

	err = writer.store.root.Link(writer.tmpRelPath, relPath)
	if err != nil {
		if errors.Is(err, fs.ErrExist) {
			writer.finalID = id
			cleanup = false
			_ = writer.store.root.Remove(writer.tmpRelPath)

			return id, nil
		}

		writer.finalErr = err

		return zero, err
	}

	writer.finalID = id
	cleanup = false

	return id, nil
}
