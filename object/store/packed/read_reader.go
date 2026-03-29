package packed

import (
	"bytes"
	"fmt"
	"io"

	"codeberg.org/lindenii/furgit/internal/iolimit"
	objectheader "codeberg.org/lindenii/furgit/object/header"
	objectid "codeberg.org/lindenii/furgit/object/id"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// ReadReaderContent reads an object's type, declared content size, and content
// stream.
//
// Close releases reader-local resources only. It does not drain unread data for
// additional validation. In particular, malformed trailing compressed data,
// trailing bytes past the declared object size, and the zlib Adler-32 trailer
// may go unverified unless the caller reads to io.EOF.
func (store *Store) ReadReaderContent(id objectid.ObjectID) (objecttype.Type, int64, io.ReadCloser, error) {
	loc, err := store.lookup(id)
	if err != nil {
		return objecttype.TypeInvalid, 0, nil, err
	}

	pack, meta, err := store.entryMetaAt(loc)
	if err != nil {
		return objecttype.TypeInvalid, 0, nil, err
	}

	if meta.ty.IsBaseObject() {
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
// Close releases reader-local resources only. It does not drain unread data for
// additional validation. In particular, malformed trailing compressed data,
// trailing bytes past the declared object size, and the zlib Adler-32 trailer
// may go unverified unless the caller reads to io.EOF.
func (store *Store) ReadReaderFull(id objectid.ObjectID) (io.ReadCloser, error) {
	loc, err := store.lookup(id)
	if err != nil {
		return nil, err
	}

	pack, meta, err := store.entryMetaAt(loc)
	if err != nil {
		return nil, err
	}

	if meta.ty.IsBaseObject() {
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
