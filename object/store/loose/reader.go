package loose

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"lindenii.org/go/furgit/internal/compress/zlib"
	"lindenii.org/go/furgit/internal/iolimit"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store"
	"lindenii.org/go/furgit/object/typ"
)

// ReadBytesFull reads a full serialized object as "type size\x00content".
//
// It inflates and parses the full loose object,
// including verifying the zlib Adler-32 trailer.
func (loose *Loose) ReadBytesFull(objectID id.ObjectID) ([]byte, error) {
	raw, _, _, err := loose.readBytesParsed(objectID)
	if err != nil {
		return nil, err
	}

	return raw, nil
}

// ReadBytesContent reads an object's type and content bytes.
//
// Like ReadBytesFull,
// it inflates and parses the full loose object,
// including verifying the zlib Adler-32 trailer.
func (loose *Loose) ReadBytesContent(objectID id.ObjectID) (typ.Type, []byte, error) {
	_, ty, content, err := loose.readBytesParsed(objectID)
	if err != nil {
		return typ.Unknown, nil, err
	}

	return ty, content, nil
}

// ReadHeader reads an object's type and declared content length.
//
// It parses only enough of the zlib-decoded object to recover the object header.
// It does not verify that the remaining object content is readable
// and does not verify the zlib Adler-32 trailer.
func (loose *Loose) ReadHeader(objectID id.ObjectID) (typ.Type, uint64, error) {
	file, err := loose.openObject(objectID)
	if err != nil {
		return typ.Unknown, 0, err
	}

	defer func() { _ = file.Close() }()

	zr, err := zlib.NewReader(file)
	if err != nil {
		return typ.Unknown, 0, fmt.Errorf("object/store/loose: %w", err)
	}

	defer func() { _ = zr.Close() }()

	_, ty, size, err := readHeader(bufio.NewReader(zr))
	if err != nil {
		return typ.Unknown, 0, err
	}

	return ty, size, nil
}

// ReadSize reads an object's declared content length.
//
// Like ReadHeader,
// it parses only enough of the zlib-decoded object to recover the header
// and does not verify the zlib Adler-32 trailer.
func (loose *Loose) ReadSize(objectID id.ObjectID) (uint64, error) {
	_, size, err := loose.ReadHeader(objectID)

	return size, err
}

// ReadReaderFull reads a full serialized object stream as "type size\x00content".
//
// Close releases resources only.
// It does not drain unread data for additional validation.
// In particular,
// malformed trailing compressed data,
// trailing bytes past the declared object size,
// and the zlib Adler-32 trailer
// may go unverified unless the caller reads to io.EOF.
func (loose *Loose) ReadReaderFull(objectID id.ObjectID) (io.ReadCloser, error) {
	file, zr, err := loose.openInflated(objectID)
	if err != nil {
		return nil, err
	}

	br := bufio.NewReader(zr)

	headerBytes, _, size, err := readHeader(br)
	if err != nil {
		_ = zr.Close()
		_ = file.Close()

		return nil, err
	}

	return &objectReader{
		reader: io.MultiReader(
			bytes.NewReader(headerBytes),
			iolimit.ExpectLengthReader(br, size),
		),
		file: file,
		zr:   zr,
	}, nil
}

// ReadReaderContent reads an object's type, declared content length,
// and content stream.
//
// Close releases resources only.
// It does not drain unread data for additional validation.
// In particular,
// malformed trailing compressed data,
// trailing bytes past the declared object size,
// and the zlib Adler-32 trailer
// may go unverified unless the caller reads to io.EOF.
func (loose *Loose) ReadReaderContent(objectID id.ObjectID) (typ.Type, uint64, io.ReadCloser, error) {
	file, zr, err := loose.openInflated(objectID)
	if err != nil {
		return typ.Unknown, 0, nil, err
	}

	br := bufio.NewReader(zr)

	_, ty, size, err := readHeader(br)
	if err != nil {
		_ = zr.Close()
		_ = file.Close()

		return typ.Unknown, 0, nil, err
	}

	return ty, size, &objectReader{
		reader: iolimit.ExpectLengthReader(br, size),
		file:   file,
		zr:     zr,
	}, nil
}

// Refresh is a no-op for loose object stores.
func (loose *Loose) Refresh() error {
	return nil
}

// openObject opens the loose object file for objectID.
// Missing files cause store.ErrObjectNotFound.
func (loose *Loose) openObject(objectID id.ObjectID) (*os.File, error) {
	relPath, err := loose.objectPath(objectID)
	if err != nil {
		return nil, err
	}

	file, err := loose.root.Open(relPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, store.ErrObjectNotFound
		}

		return nil, fmt.Errorf("object/store/loose: %w", err)
	}

	return file, nil
}

// readBytesParsed reads, inflates, and parses a loose object in one pass.
// It returns the full raw payload and its parsed type and content.
func (loose *Loose) readBytesParsed(objectID id.ObjectID) ([]byte, typ.Type, []byte, error) {
	file, err := loose.openObject(objectID)
	if err != nil {
		return nil, typ.Unknown, nil, err
	}

	defer func() { _ = file.Close() }()

	raw, err := decodeAll(file)
	if err != nil {
		return nil, typ.Unknown, nil, err
	}

	ty, content, err := parseRaw(raw)
	if err != nil {
		return nil, typ.Unknown, nil, err
	}

	return raw, ty, content, nil
}

// openInflated opens and zlib-decodes a loose object file.
// The caller owns both returned closers and must close them.
func (loose *Loose) openInflated(objectID id.ObjectID) (*os.File, io.ReadCloser, error) {
	file, err := loose.openObject(objectID)
	if err != nil {
		return nil, nil, err
	}

	zr, err := zlib.NewReader(file)
	if err != nil {
		_ = file.Close()

		return nil, nil, fmt.Errorf("object/store/loose: %w", err)
	}

	return file, zr, nil
}

// objectReader streams one inflated loose object
// and owns the underlying file and zlib decoder.
type objectReader struct {
	// reader is the stream exposed by Read.
	reader io.Reader
	// file is the underlying loose object file and is closed by Close.
	file *os.File
	// zr is the zlib decoder and is closed by Close.
	zr io.ReadCloser
}

func (reader *objectReader) Read(dst []byte) (int, error) {
	return reader.reader.Read(dst) //nolint:wrapcheck
}

func (reader *objectReader) Close() error {
	errZlib := reader.zr.Close()
	errFile := reader.file.Close()

	return errors.Join(errZlib, errFile)
}
