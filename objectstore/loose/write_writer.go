package loose

import (
	"errors"
	"hash"
	"os"

	"codeberg.org/lindenii/furgit/internal/compress/zlib"
	"codeberg.org/lindenii/furgit/objectid"
)

const tempObjectFilePrefix = "tmp_obj_"

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
		err := writer.acceptFull(src)
		if err != nil {
			return 0, err
		}
	} else {
		err := writer.acceptContent(int64(len(src)))
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
