package memory

import (
	"errors"
	"fmt"
	"io"

	objectheader "codeberg.org/lindenii/furgit/object/header"
	objectid "codeberg.org/lindenii/furgit/object/id"
	objecttype "codeberg.org/lindenii/furgit/object/type"
)

// WriteReaderContent writes one typed object content stream.
func (store *Store) WriteReaderContent(ty objecttype.Type, size int64, src io.Reader) (objectid.ObjectID, error) {
	if size < 0 {
		return objectid.ObjectID{}, fmt.Errorf("objectstore/memory: negative content size: %d", size)
	}

	content, err := io.ReadAll(io.LimitReader(src, size+1))
	if err != nil {
		return objectid.ObjectID{}, err
	}

	switch {
	case int64(len(content)) > size:
		return objectid.ObjectID{}, errors.New("objectstore/memory: object content longer than declared size")
	case int64(len(content)) < size:
		return objectid.ObjectID{}, errors.New("objectstore/memory: object content shorter than declared size")
	}

	return store.WriteBytesContent(ty, content)
}

// WriteReaderFull writes one full serialized object stream as "type size\0content".
func (store *Store) WriteReaderFull(src io.Reader) (objectid.ObjectID, error) {
	raw, err := io.ReadAll(src)
	if err != nil {
		return objectid.ObjectID{}, err
	}

	ty, size, headerLen, ok := objectheader.Parse(raw)
	if !ok {
		return objectid.ObjectID{}, errors.New("objectstore/memory: malformed object header")
	}

	content := raw[headerLen:]
	if int64(len(content)) != size {
		return objectid.ObjectID{}, errors.New("objectstore/memory: object header size/content mismatch")
	}

	id := store.algo.Sum(raw)
	store.objects[id] = storedObject{ty: ty, content: append([]byte(nil), content...)}

	return id, nil
}
