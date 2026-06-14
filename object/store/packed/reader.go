package packed

import (
	"bytes"
	"fmt"
	"io"

	"lindenii.org/go/furgit/internal/compress/zlib"
	"lindenii.org/go/furgit/internal/format/packfile"
	"lindenii.org/go/furgit/internal/iolimit"
	"lindenii.org/go/furgit/object/header"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/typ"
	"lindenii.org/go/lgo/intconv"
)

// ReadBytesContent reads an object's type and content bytes,
// fully resolving delta chains.
func (packed *Packed) ReadBytesContent(objectID id.ObjectID) (typ.Type, []byte, error) {
	p, offset, err := packed.lookup(objectID)
	if err != nil {
		return typ.Unknown, nil, err
	}

	entryType, content, err := packed.unpackEntry(p, offset)
	if err != nil {
		return typ.Unknown, nil, err
	}

	ty, err := objectTypeOf(entryType)
	if err != nil {
		return typ.Unknown, nil, err
	}

	return ty, content, nil
}

// ReadBytesFull reads a full serialized object as "type size\x00content",
// fully resolving delta chains.
func (packed *Packed) ReadBytesFull(objectID id.ObjectID) ([]byte, error) {
	ty, content, err := packed.ReadBytesContent(objectID)
	if err != nil {
		return nil, err
	}

	raw := header.Append(make([]byte, 0, len(content)+32), ty, len(content))

	return append(raw, content...), nil
}

// ReadHeader reads an object's type and declared content length.
//
// For delta entries this resolves the chained base type
// and the declared delta result size,
// without reconstructing content.
func (packed *Packed) ReadHeader(objectID id.ObjectID) (typ.Type, int, error) {
	p, offset, err := packed.lookup(objectID)
	if err != nil {
		return typ.Unknown, 0, err
	}

	entryHeader, payload, err := p.entryHeaderAt(offset, packed.objectFormat)
	if err != nil {
		return typ.Unknown, 0, err
	}

	var size int

	if entryHeader.Type.IsDelta() {
		size, err = deltaResultSize(payload, entryHeader.Size)
		if err != nil {
			return typ.Unknown, 0, fmt.Errorf("%w: pack %q: %w", ErrMalformedPackedStore, p.name, err)
		}
	} else {
		size, err = intconv.Uint64ToInt(entryHeader.Size)
		if err != nil {
			return typ.Unknown, 0, fmt.Errorf("%w: pack %q: object size overflows int: %w", ErrMalformedPackedStore, p.name, err)
		}
	}

	entryType, err := packed.resolveType(p, offset, entryHeader)
	if err != nil {
		return typ.Unknown, 0, err
	}

	ty, err := objectTypeOf(entryType)
	if err != nil {
		return typ.Unknown, 0, err
	}

	return ty, size, nil
}

// ReadSize reads an object's declared content length.
//
// Unlike ReadHeader,
// this never walks delta chains.
func (packed *Packed) ReadSize(objectID id.ObjectID) (int, error) {
	p, offset, err := packed.lookup(objectID)
	if err != nil {
		return 0, err
	}

	entryHeader, payload, err := p.entryHeaderAt(offset, packed.objectFormat)
	if err != nil {
		return 0, err
	}

	if !entryHeader.Type.IsDelta() {
		size, err := intconv.Uint64ToInt(entryHeader.Size)
		if err != nil {
			return 0, fmt.Errorf("%w: pack %q: object size overflows int: %w", ErrMalformedPackedStore, p.name, err)
		}

		return size, nil
	}

	size, err := deltaResultSize(payload, entryHeader.Size)
	if err != nil {
		return 0, fmt.Errorf("%w: pack %q: %w", ErrMalformedPackedStore, p.name, err)
	}

	return size, nil
}

// ReadReaderContent reads an object's type,
// declared content length, and content stream.
//
// Non-delta entries stream directly from the pack;
// delta entries are fully resolved in memory first.
// Close releases resources only
// and does not drain unread data for additional validation.
func (packed *Packed) ReadReaderContent(objectID id.ObjectID) (typ.Type, int, io.ReadCloser, error) {
	p, offset, err := packed.lookup(objectID)
	if err != nil {
		return typ.Unknown, 0, nil, err
	}

	entryHeader, payload, err := p.entryHeaderAt(offset, packed.objectFormat)
	if err != nil {
		return typ.Unknown, 0, nil, err
	}

	if !entryHeader.Type.IsBase() {
		entryType, content, err := packed.unpackEntry(p, offset)
		if err != nil {
			return typ.Unknown, 0, nil, err
		}

		ty, err := objectTypeOf(entryType)
		if err != nil {
			return typ.Unknown, 0, nil, err
		}

		return ty, len(content), io.NopCloser(bytes.NewReader(content)), nil
	}

	ty, err := objectTypeOf(entryHeader.Type)
	if err != nil {
		return typ.Unknown, 0, nil, err
	}

	size, err := intconv.Uint64ToInt(entryHeader.Size)
	if err != nil {
		return typ.Unknown, 0, nil, fmt.Errorf("%w: pack %q: object size overflows int: %w", ErrMalformedPackedStore, p.name, err)
	}

	zr, err := zlib.NewReaderBytes(payload)
	if err != nil {
		return typ.Unknown, 0, nil, fmt.Errorf("%w: pack %q: %w", ErrMalformedPackedStore, p.name, err)
	}

	return ty, size, &objectReader{
		reader: iolimit.ExpectLengthReader(zr, size),
		zr:     zr,
	}, nil
}

// ReadReaderFull reads a full serialized object stream
// as "type size\x00content".
//
// Non-delta entries stream directly from the pack;
// delta entries are fully resolved in memory first.
// Close releases resources only
// and does not drain unread data for additional validation.
func (packed *Packed) ReadReaderFull(objectID id.ObjectID) (io.ReadCloser, error) {
	ty, size, reader, err := packed.ReadReaderContent(objectID)
	if err != nil {
		return nil, err
	}

	headerBytes := header.Append(nil, ty, size)

	return &objectReader{
		reader: io.MultiReader(bytes.NewReader(headerBytes), reader),
		zr:     reader,
	}, nil
}

// objectTypeOf converts one packfile entry type
// into an ordinary object type.
func objectTypeOf(entryType packfile.EntryType) (typ.Type, error) {
	ty, err := entryType.ObjectType()
	if err != nil {
		return typ.Unknown, fmt.Errorf("%w: %w", ErrMalformedPackedStore, err)
	}

	return ty, nil
}

// objectReader streams one packed object payload
// and owns the underlying decompressor.
type objectReader struct {
	// reader is the stream exposed by Read.
	reader io.Reader
	// zr is closed by Close.
	zr io.Closer
}

func (reader *objectReader) Read(dst []byte) (int, error) {
	return reader.reader.Read(dst)
}

func (reader *objectReader) Close() error {
	return reader.zr.Close()
}
