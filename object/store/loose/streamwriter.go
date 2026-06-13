package loose

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"hash"
	"io/fs"
	"os"
	"path/filepath"

	"lindenii.org/go/furgit/internal/compress/zlib"
	"lindenii.org/go/furgit/object/header"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store"
)

const tempObjectFilePrefix = "tmp_obj_"

// streamWriter incrementally hashes and deflates an object into a temp file.
// finalize validates size accounting and atomically renames the temp file.
type streamWriter struct {
	// loose owns path and root operations used by this write session.
	loose *Loose
	// file is the temporary destination file under objects/.
	file *os.File
	// zw compresses raw object bytes into file.
	zw *zlib.Writer
	// hash receives the same raw bytes used to compute the resulting object ID.
	hash hash.Hash

	// tmpRelPath is the relative path of file under the objects root.
	tmpRelPath string

	// fullMode selects full-object input ("type size\x00content")
	// as opposed to content-only input.
	fullMode bool

	// headerBuf accumulates header bytes while fullMode parses up to the first NUL.
	headerBuf []byte
	// headerDone reports whether the full-object header has been parsed.
	headerDone bool
	// expectedContentLeft tracks remaining declared content bytes.
	expectedContentLeft int

	closed    bool
	finalized bool
}

// Write validates and writes raw bytes into the stream.
// In full mode, it parses and enforces the streamed header-declared content size.
func (writer *streamWriter) Write(src []byte) (int, error) {
	if writer.finalized {
		return 0, fmt.Errorf("%w: write after finalize", store.ErrInvalidObject)
	}

	if writer.closed {
		return 0, fmt.Errorf("%w: write after close", store.ErrInvalidObject)
	}

	if writer.fullMode {
		err := writer.acceptFull(src)
		if err != nil {
			return 0, err
		}
	} else {
		err := writer.acceptContent(len(src))
		if err != nil {
			return 0, err
		}
	}

	err := writer.writeRawChunk(src)
	if err != nil {
		return 0, err
	}

	return len(src), nil
}

// Close flushes and closes the underlying zlib stream and temp file.
func (writer *streamWriter) Close() error {
	errZlib := writer.zw.Close()
	errSync := writer.file.Sync()
	errFile := writer.file.Close()

	writer.closed = true
	writer.file = nil

	return errors.Join(errZlib, errSync, errFile)
}

// acceptFull validates and accounts raw full-object input.
func (writer *streamWriter) acceptFull(src []byte) error {
	if writer.headerDone {
		return writer.acceptContent(len(src))
	}

	nul := bytes.IndexByte(src, 0)
	if nul < 0 {
		writer.headerBuf = append(writer.headerBuf, src...)

		return nil
	}

	headerChunkLen := nul + 1
	writer.headerBuf = append(writer.headerBuf, src[:headerChunkLen]...)

	_, size, _, err := header.Parse(writer.headerBuf)
	if err != nil {
		return fmt.Errorf("%w: %w", store.ErrInvalidObject, err)
	}

	writer.headerDone = true
	writer.expectedContentLeft = size

	return writer.acceptContent(len(src) - headerChunkLen)
}

// acceptContent validates and accounts content byte counts.
func (writer *streamWriter) acceptContent(n int) error {
	if n > writer.expectedContentLeft {
		return fmt.Errorf("%w: object content exceeds declared size", store.ErrInvalidObject)
	}

	writer.expectedContentLeft -= n

	return nil
}

// writeRawChunk forwards raw bytes to the hash and deflate pipeline.
func (writer *streamWriter) writeRawChunk(src []byte) error {
	_, err := writer.hash.Write(src)
	if err != nil {
		return fmt.Errorf("object/store/loose: %w", err)
	}

	_, err = writer.zw.Write(src)
	if err != nil {
		return fmt.Errorf("object/store/loose: %w", err)
	}

	return nil
}

// finalize validates write completeness and atomically publishes the object.
// Publication is no-clobber: it links tmpRelPath to the object path and treats
// existing destination objects as success.
func (writer *streamWriter) finalize() (id.ObjectID, error) {
	writer.finalized = true

	var zero id.ObjectID

	if !writer.closed {
		err := writer.Close()
		if err != nil {
			return zero, err
		}
	}

	if writer.fullMode && !writer.headerDone {
		return zero, fmt.Errorf("%w: missing full object header", store.ErrInvalidObject)
	}

	if writer.expectedContentLeft != 0 {
		return zero, fmt.Errorf("%w: object content shorter than declared size", store.ErrInvalidObject)
	}

	idBytes := writer.hash.Sum(nil)

	objectID, err := writer.loose.objectFormat.FromBytes(idBytes)
	if err != nil {
		return zero, fmt.Errorf("object/store/loose: %w", err)
	}

	relPath, err := writer.loose.objectPath(objectID)
	if err != nil {
		return zero, err
	}

	dir := filepath.Dir(relPath)

	err = writer.loose.root.MkdirAll(dir, 0o755)
	if err != nil {
		return zero, fmt.Errorf("object/store/loose: %w", err)
	}

	cleanup := true

	defer func() {
		if cleanup {
			_ = writer.loose.root.Remove(writer.tmpRelPath)
		}
	}()

	err = writer.loose.root.Link(writer.tmpRelPath, relPath)
	if err != nil {
		if errors.Is(err, fs.ErrExist) {
			cleanup = false
			_ = writer.loose.root.Remove(writer.tmpRelPath)

			return objectID, nil
		}

		return zero, fmt.Errorf("object/store/loose: %w", err)
	}

	cleanup = false
	_ = writer.loose.root.Remove(writer.tmpRelPath)

	return objectID, nil
}

// newStreamWriter creates a stream writer with a temp file rooted in objects/.
func (loose *Loose) newStreamWriter(fullMode bool) (*streamWriter, error) {
	hashFn, err := loose.objectFormat.New()
	if err != nil {
		return nil, fmt.Errorf("object/store/loose: %w", err)
	}

	tmpRelPath, file, err := loose.createTempObjectFile(".")
	if err != nil {
		return nil, err
	}

	return &streamWriter{
		loose:      loose,
		file:       file,
		zw:         zlib.NewWriter(file),
		hash:       hashFn,
		tmpRelPath: tmpRelPath,
		fullMode:   fullMode,
		headerBuf:  make([]byte, 0, 64),
	}, nil
}

// createTempObjectFile creates a unique temporary object file within dir.
// The returned path is relative to the objects root.
func (loose *Loose) createTempObjectFile(dir string) (string, *os.File, error) {
	var lastErr error

	for range 16 {
		relPath := filepath.Join(dir, tempObjectFilePrefix+rand.Text())

		file, err := loose.root.OpenFile(relPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			return relPath, file, nil
		}

		lastErr = err

		if errors.Is(err, fs.ErrExist) {
			continue
		}

		return "", nil, fmt.Errorf("object/store/loose: %w", err)
	}

	return "", nil, fmt.Errorf("object/store/loose: failed to create temporary object file: %w", lastErr)
}
