package loose

import (
	"errors"
	"io/fs"
	"path/filepath"

	objectid "lindenii.org/go/furgit/object/id"
)

// Close flushes and closes the underlying zlib stream and temp file.
func (writer *streamWriter) Close() error {
	errZlib := writer.zw.Close()
	errSync := writer.file.Sync()
	errFile := writer.file.Close()

	writer.closed = true
	writer.file = nil

	return errors.Join(errZlib, errSync, errFile)
}

// finalize validates write completeness and atomically publishes the object.
// Publication is no-clobber: it links tmpRelPath to the object path and treats
// existing destination objects as success.
func (writer *streamWriter) finalize() (objectid.ObjectID, error) {
	writer.finalized = true

	var zero objectid.ObjectID

	if !writer.closed {
		err := writer.Close()
		if err != nil {
			return zero, err
		}
	}

	if writer.fullMode && !writer.headerDone {
		return zero, errors.New("objectstore/loose: missing full object header")
	}

	if writer.expectedContentLeft != 0 {
		return zero, errors.New("objectstore/loose: object content shorter than declared size")
	}

	idBytes := writer.hash.Sum(nil)

	id, err := objectid.FromBytes(writer.store.algo, idBytes)
	if err != nil {
		return zero, err
	}

	relPath, err := writer.store.objectPath(id)
	if err != nil {
		return zero, err
	}

	dir := filepath.Dir(relPath)

	err = writer.store.root.MkdirAll(dir, 0o755)
	if err != nil {
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
			cleanup = false
			_ = writer.store.root.Remove(writer.tmpRelPath)

			return id, nil
		}

		return zero, err
	}

	cleanup = false
	_ = writer.store.root.Remove(writer.tmpRelPath)

	return id, nil
}
