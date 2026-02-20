package loose

import (
	"bytes"
	"compress/zlib"
	"crypto/rand"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"codeberg.org/lindenii/furgit/objectheader"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
)

const tempObjectFilePrefix = "tmp_obj_"

// WriteWriterContent returns a writer for object content bytes.
// The writer accepts exactly size bytes. After closing the writer,
// call finalize to atomically publish the loose object and get its ID.
func (store *Store) WriteWriterContent(ty objecttype.Type, size int64) (io.WriteCloser, func() (objectid.ObjectID, error), error) {
	if size < 0 {
		return nil, nil, errors.New("objectstore/loose: negative content size")
	}

	header, ok := objectheader.Encode(ty, size)
	if !ok {
		return nil, nil, fmt.Errorf("objectstore/loose: failed to encode object header for type %d", ty)
	}

	writer, err := store.newStreamWriter(false)
	if err != nil {
		return nil, nil, err
	}
	writer.headerDone = true
	writer.expectedContentLeft = size
	if err := writer.writeRawChunk(header); err != nil {
		_ = writer.Close()
		return nil, nil, err
	}

	return writer, writer.Finalize, nil
}

// WriteWriterFull returns a writer for full raw object bytes:
// "type size\0content". After closing the writer, call finalize
// to atomically publish the loose object and get its ID.
func (store *Store) WriteWriterFull() (io.WriteCloser, func() (objectid.ObjectID, error), error) {
	writer, err := store.newStreamWriter(true)
	if err != nil {
		return nil, nil, err
	}
	return writer, writer.Finalize, nil
}

// streamWriter incrementally hashes and deflates an object into a temp file.
// Finalize validates size accounting and atomically renames the temp file.
type streamWriter struct {
	// store owns path and root operations used by this write session.
	store *Store
	// file is the temporary destination file under objects/.
	file *os.File
	// zw compresses raw object bytes into file.
	zw *zlib.Writer
	// hash receives the same raw bytes used to compute the resulting object ID.
	hash hash.Hash

	// tmpRelPath is the relative path of file under the objects root.
	tmpRelPath string

	// fullMode selects full-object input ("type size\0content") as opposed to content-only input.
	fullMode bool

	// headerBuf accumulates header bytes while fullMode parses up to the first NUL.
	headerBuf []byte
	// headerDone reports whether the full-object header has been parsed.
	headerDone bool
	// expectedContentLeft tracks remaining declared content bytes.
	expectedContentLeft int64

	closed    bool
	finalized bool
	finalID   objectid.ObjectID
	finalErr  error
}

// newStreamWriter creates a stream writer with a temp file rooted in objects/.
func (store *Store) newStreamWriter(fullMode bool) (*streamWriter, error) {
	hashFn, err := store.algo.New()
	if err != nil {
		return nil, err
	}

	tmpRelPath, file, err := store.createTempObjectFile(".")
	if err != nil {
		return nil, err
	}

	return &streamWriter{
		store:      store,
		file:       file,
		zw:         zlib.NewWriter(file),
		hash:       hashFn,
		tmpRelPath: tmpRelPath,
		fullMode:   fullMode,
		headerBuf:  make([]byte, 0, 64),
	}, nil
}

// Write validates and writes raw bytes into the stream.
// In full mode, it parses and enforces the streamed header-declared content size.
func (writer *streamWriter) Write(src []byte) (int, error) {
	if writer.finalized {
		return 0, errors.New("objectstore/loose: write after finalize")
	}
	if writer.closed {
		return 0, errors.New("objectstore/loose: write after close")
	}

	if writer.fullMode {
		if err := writer.acceptFull(src); err != nil {
			return 0, err
		}
	} else {
		if err := writer.acceptContent(int64(len(src))); err != nil {
			return 0, err
		}
	}

	if err := writer.writeRawChunk(src); err != nil {
		return 0, err
	}
	return len(src), nil
}

// acceptFull validates and accounts raw full-object input.
func (writer *streamWriter) acceptFull(src []byte) error {
	if !writer.headerDone {
		if nul := bytes.IndexByte(src, 0); nul >= 0 {
			headerChunkLen := nul + 1
			writer.headerBuf = append(writer.headerBuf, src[:headerChunkLen]...)
			_, size, _, ok := objectheader.Parse(writer.headerBuf)
			if !ok {
				return errors.New("objectstore/loose: malformed object header")
			}
			writer.headerDone = true
			writer.expectedContentLeft = size
			return writer.acceptContent(int64(len(src) - headerChunkLen))
		}

		writer.headerBuf = append(writer.headerBuf, src...)
		return nil
	}

	return writer.acceptContent(int64(len(src)))
}

// acceptContent validates and accounts content byte counts.
func (writer *streamWriter) acceptContent(n int64) error {
	if n > writer.expectedContentLeft {
		return errors.New("objectstore/loose: object content exceeds declared size")
	}
	writer.expectedContentLeft -= n
	return nil
}

// writeRawChunk forwards raw bytes to the hash and deflate pipeline.
func (writer *streamWriter) writeRawChunk(src []byte) error {
	if _, err := writer.hash.Write(src); err != nil {
		return err
	}
	if _, err := writer.zw.Write(src); err != nil {
		return err
	}
	return nil
}

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

// Finalize validates write completeness and atomically publishes the object.
// Publication is no-clobber: it links tmpRelPath to the object path and treats
// existing destination objects as success.
func (writer *streamWriter) Finalize() (objectid.ObjectID, error) {
	if writer.finalized {
		return writer.finalID, writer.finalErr
	}
	writer.finalized = true

	var zero objectid.ObjectID

	if !writer.closed {
		if err := writer.Close(); err != nil {
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
	if err := writer.store.root.MkdirAll(dir, 0o755); err != nil {
		writer.finalErr = err
		return zero, err
	}

	cleanup := true
	defer func() {
		if cleanup {
			_ = writer.store.root.Remove(writer.tmpRelPath)
		}
	}()

	if err := writer.store.root.Link(writer.tmpRelPath, relPath); err != nil {
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

// createTempObjectFile creates a unique temporary object file within dir.
// The returned path is relative to the objects root.
func (store *Store) createTempObjectFile(dir string) (string, *os.File, error) {
	for range 16 {
		relPath := filepath.Join(dir, tempObjectFilePrefix+rand.Text())
		file, err := store.root.OpenFile(relPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			return relPath, file, nil
		}
		if errors.Is(err, fs.ErrExist) {
			continue
		}
		return "", nil, err
	}

	return "", nil, errors.New("objectstore/loose: failed to create temporary object file")
}
