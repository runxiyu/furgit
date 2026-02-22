package loose

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"

	"codeberg.org/lindenii/furgit/internal/iolimit"
	"codeberg.org/lindenii/furgit/internal/zlib"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
)

type objectReader struct {
	// reader is the stream exposed by Read.
	reader io.Reader
	// file is the underlying loose object file and is closed by Close.
	file *os.File
	// zr is the zlib decoder and is closed by Close.
	zr io.ReadCloser
}

func (reader *objectReader) Read(dst []byte) (int, error) {
	return reader.reader.Read(dst)
}

func (reader *objectReader) Close() error {
	errZlib := reader.zr.Close()
	errFile := reader.file.Close()
	return errors.Join(errZlib, errFile)
}

// openInflated opens and zlib-decodes a loose object file.
// The caller owns both returned closers and must close them.
func (store *Store) openInflated(id objectid.ObjectID) (*os.File, io.ReadCloser, error) {
	file, err := store.openObject(id)
	if err != nil {
		return nil, nil, err
	}
	zr, err := zlib.NewReader(file)
	if err != nil {
		_ = file.Close()
		return nil, nil, err
	}
	return file, zr, nil
}

// ReadReaderFull reads a full serialized object stream as "type size\0content".
// The caller must close the returned reader.
func (store *Store) ReadReaderFull(id objectid.ObjectID) (io.ReadCloser, error) {
	file, zr, err := store.openInflated(id)
	if err != nil {
		return nil, err
	}

	br := bufio.NewReader(zr)
	header, _, size, err := readHeader(br)
	if err != nil {
		_ = zr.Close()
		_ = file.Close()
		return nil, err
	}

	return &objectReader{
		reader: io.MultiReader(
			bytes.NewReader(header),
			iolimit.ExpectLengthReader(br, size),
		),
		file: file,
		zr:   zr,
	}, nil
}

// ReadReaderContent reads an object's type, declared content length, and content stream.
// The caller must close the returned reader.
func (store *Store) ReadReaderContent(id objectid.ObjectID) (objecttype.Type, int64, io.ReadCloser, error) {
	file, zr, err := store.openInflated(id)
	if err != nil {
		return objecttype.TypeInvalid, 0, nil, err
	}

	br := bufio.NewReader(zr)
	_, ty, size, err := readHeader(br)
	if err != nil {
		_ = zr.Close()
		_ = file.Close()
		return objecttype.TypeInvalid, 0, nil, err
	}

	return ty, size, &objectReader{
		reader: iolimit.ExpectLengthReader(br, size),
		file:   file,
		zr:     zr,
	}, nil
}
