package memory

import (
	"fmt"
	"io"

	"lindenii.org/go/furgit/object/header"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store"
	"lindenii.org/go/furgit/object/typ"
)

// WriteBytesContent writes one typed object content byte slice.
func (memory *Memory) WriteBytesContent(ty typ.Type, content []byte) (id.ObjectID, error) {
	raw := header.Append(nil, ty, len(content))
	raw = append(raw, content...)

	objectID := memory.objectFormat.Sum(raw)
	memory.objects.Store(objectID, storedObject{ty: ty, content: append([]byte(nil), content...)})

	return objectID, nil
}

// WriteBytesFull writes one full serialized object byte slice as "type size\x00content".
func (memory *Memory) WriteBytesFull(raw []byte) (id.ObjectID, error) {
	ty, size, consumed, err := header.Parse(raw)
	if err != nil {
		return id.ObjectID{}, fmt.Errorf("object/store/memory: %w", err)
	}

	content := raw[consumed:]
	if len(content) != size {
		return id.ObjectID{}, fmt.Errorf("%w: header size/content mismatch", store.ErrInvalidObject)
	}

	return memory.WriteBytesContent(ty, content)
}

// WriteReaderContent writes one typed object content stream.
func (memory *Memory) WriteReaderContent(ty typ.Type, size int, src io.Reader) (id.ObjectID, error) {
	content, err := io.ReadAll(io.LimitReader(src, int64(size)+1))
	if err != nil {
		return id.ObjectID{}, fmt.Errorf("object/store/memory: read content: %w", err)
	}

	switch {
	case len(content) > size:
		return id.ObjectID{}, fmt.Errorf("%w: content longer than declared size", store.ErrInvalidObject)
	case len(content) < size:
		return id.ObjectID{}, fmt.Errorf("%w: content shorter than declared size", store.ErrInvalidObject)
	}

	return memory.WriteBytesContent(ty, content)
}

// WriteReaderFull writes one full serialized object stream as "type size\x00content".
func (memory *Memory) WriteReaderFull(src io.Reader) (id.ObjectID, error) {
	raw, err := io.ReadAll(src)
	if err != nil {
		return id.ObjectID{}, fmt.Errorf("object/store/memory: read object: %w", err)
	}

	return memory.WriteBytesFull(raw)
}
