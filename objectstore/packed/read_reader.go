package packed

import (
	"bytes"
	"fmt"
	"io"

	"codeberg.org/lindenii/furgit/internal/iolimit"
	"codeberg.org/lindenii/furgit/objectheader"
	"codeberg.org/lindenii/furgit/objectid"
	"codeberg.org/lindenii/furgit/objecttype"
)

// readCloser proxies reads and closes one underlying closer.
type readCloser struct {
	reader io.Reader
	closer io.Closer
}

// Read proxies reads to the underlying reader.
func (reader *readCloser) Read(dst []byte) (int, error) {
	return reader.reader.Read(dst)
}

// Close closes the underlying closer.
func (reader *readCloser) Close() error {
	return reader.closer.Close()
}

// ReadReaderContent reads an object's type, declared content size, and content stream.
//
// The caller must close the returned reader.
func (store *Store) ReadReaderContent(id objectid.ObjectID) (objecttype.Type, int64, io.ReadCloser, error) {
	loc, err := store.lookup(id)
	if err != nil {
		return objecttype.TypeInvalid, 0, nil, err
	}

	pack, meta, err := store.entryMetaAt(loc)
	if err != nil {
		return objecttype.TypeInvalid, 0, nil, err
	}
	if isBaseObjectType(meta.ty) {
		zr, err := zlibReaderAt(pack, meta.dataOffset)
		if err != nil {
			return objecttype.TypeInvalid, 0, nil, err
		}
		return meta.ty, meta.size, &readCloser{
			reader: iolimit.ExpectLengthReader(zr, meta.size),
			closer: zr,
		}, nil
	}

	ty, content, err := store.deltaResolveContent(loc)
	if err != nil {
		return objecttype.TypeInvalid, 0, nil, err
	}
	return ty, int64(len(content)), io.NopCloser(bytes.NewReader(content)), nil
}

// ReadReaderFull reads a full serialized object stream as "type size\0content".
//
// The caller must close the returned reader.
func (store *Store) ReadReaderFull(id objectid.ObjectID) (io.ReadCloser, error) {
	loc, err := store.lookup(id)
	if err != nil {
		return nil, err
	}

	pack, meta, err := store.entryMetaAt(loc)
	if err != nil {
		return nil, err
	}
	if isBaseObjectType(meta.ty) {
		header, ok := objectheader.Encode(meta.ty, meta.size)
		if !ok {
			return nil, fmt.Errorf("objectstore/packed: failed to encode object header for type %d", meta.ty)
		}
		zr, err := zlibReaderAt(pack, meta.dataOffset)
		if err != nil {
			return nil, err
		}
		return &readCloser{
			reader: io.MultiReader(bytes.NewReader(header), iolimit.ExpectLengthReader(zr, meta.size)),
			closer: zr,
		}, nil
	}

	raw, err := store.ReadBytesFull(id)
	if err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewReader(raw)), nil
}
