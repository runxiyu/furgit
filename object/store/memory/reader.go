package memory

import (
	"bytes"
	"io"

	"lindenii.org/go/furgit/object/header"
	"lindenii.org/go/furgit/object/id"
	"lindenii.org/go/furgit/object/store"
	"lindenii.org/go/furgit/object/typ"
)

// ReadBytesFull reads one full object, including the object header.
func (memory *Memory) ReadBytesFull(id id.ObjectID) ([]byte, error) {
	obj, ok := memory.objects.Load(id)
	if !ok {
		return nil, store.ErrObjectNotFound
	}

	raw := header.Append(nil, obj.ty, uint64(len(obj.content)))
	raw = append(raw, obj.content...)

	return raw, nil
}

// ReadBytesContent reads one object body.
func (memory *Memory) ReadBytesContent(id id.ObjectID) (typ.Type, []byte, error) {
	obj, ok := memory.objects.Load(id)
	if !ok {
		return typ.Unknown, nil, store.ErrObjectNotFound
	}

	return obj.ty, append([]byte(nil), obj.content...), nil
}

// ReadHeader reads one object header.
func (memory *Memory) ReadHeader(id id.ObjectID) (typ.Type, uint64, error) {
	obj, ok := memory.objects.Load(id)
	if !ok {
		return typ.Unknown, 0, store.ErrObjectNotFound
	}

	return obj.ty, uint64(len(obj.content)), nil
}

// ReadSize reads one object size.
func (memory *Memory) ReadSize(id id.ObjectID) (uint64, error) {
	_, size, err := memory.ReadHeader(id)
	if err != nil {
		return 0, err
	}

	return size, nil
}

// ReadReaderFull reads one full object through a reader.
func (memory *Memory) ReadReaderFull(id id.ObjectID) (io.ReadCloser, error) {
	raw, err := memory.ReadBytesFull(id)
	if err != nil {
		return nil, err
	}

	return io.NopCloser(bytes.NewReader(raw)), nil
}

// ReadReaderContent reads one object body through a reader.
func (memory *Memory) ReadReaderContent(id id.ObjectID) (typ.Type, uint64, io.ReadCloser, error) {
	ty, content, err := memory.ReadBytesContent(id)
	if err != nil {
		return typ.Unknown, 0, nil, err
	}

	return ty, uint64(len(content)), io.NopCloser(bytes.NewReader(content)), nil
}

// Refresh is a no-op for in-memory object stores.
func (memory *Memory) Refresh() error {
	return nil
}
